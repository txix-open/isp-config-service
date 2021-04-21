package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/bootstrap"
	"github.com/integration-system/isp-lib/v2/config"
	"github.com/integration-system/isp-lib/v2/metric"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	mux "github.com/integration-system/net-mux"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/conf"
	"isp-config-service/controller"
	_ "isp-config-service/docs"
	"isp-config-service/helper"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/raft"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"
	"isp-config-service/subs"
)

const (
	moduleName                         = "config"
	defaultWsConnectionReadLimit int64 = 4 << 20 // 4 MB
)

var (
	version = "0.1.0"

	muxer mux.Mux
)

func init() {
	config.InitConfig(&conf.Configuration{})
	if utils.LOG_LEVEL != "" {
		_ = log.SetLevel(utils.LOG_LEVEL)
	}
}

// @title ISP configuration service
// @version 2.4.1
// @description Сервис управления конфигурацией модулей ISP кластера

// @license.name GNU GPL v3.0

// @host localhost:9003
// @BasePath /api/config

//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json
func main() {
	cfg := config.Get().(*conf.Configuration)
	handlers := helper.GetHandlers()
	endpoints := backend.GetEndpoints(cfg.ModuleName, handlers)
	address := cfg.GrpcOuterAddress
	if address.IP == "" {
		ip, err := getOutboundIP()
		if err != nil {
			panic(err)
		}
		address.IP = ip
	}
	declaration := structure.BackendDeclaration{
		ModuleName: cfg.ModuleName,
		Version:    version,
		LibVersion: bootstrap.LibraryVersion,
		Endpoints:  endpoints,
		Address:    address,
	}

	connectionReadLimit := defaultWsConnectionReadLimit
	if cfg.WS.WsConnectionReadLimitKB > 0 {
		//nolint:gomnd
		connectionReadLimit = cfg.WS.WsConnectionReadLimitKB << 10
	}
	cluster.WsConnectionReadLimit = connectionReadLimit

	cfg.Database.Password = fmt.Sprintf("%s%s", cfg.Database.Password, os.Getenv("DB_PASSPART"))
	model.DbClient.ReceiveConfiguration(cfg.Database)

	httpListener, raftListener, err := initMultiplexer(cfg.WS.Rest)
	if err != nil {
		log.Fatalf(codes.InitMuxError, "init mux: %v", err)
	}

	_, raftStore := initRaft(raftListener, cfg.Cluster, declaration)
	initWebsocket(connectionReadLimit, httpListener, raftStore)
	initGrpc(cfg.WS.Grpc, raftStore)

	metric.InitProfiling(cfg.ModuleName)
	metric.InitCollectors(cfg.Metrics, structure.MetricConfiguration{})
	metric.InitHttpServer(cfg.Metrics)

	gracefulShutdown()
}

func initMultiplexer(addressConfiguration structure.AddressConfiguration) (net.Listener, net.Listener, error) {
	outerAddr, err := net.ResolveTCPAddr("tcp4", addressConfiguration.GetAddress())
	if err != nil {
		return nil, nil, fmt.Errorf("resolve outer address: %v", err)
	}
	tcpListener, err := net.ListenTCP("tcp4", outerAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("create tcp transport: %v", err)
	}

	muxer = mux.New(tcpListener)
	httpListener := muxer.Match(mux.HTTP1())
	raftListener := muxer.Match(mux.Any())

	go func() {
		if err := muxer.Serve(); err != nil {
			log.Fatalf(codes.InitMuxError, "serve mux: %v", err)
		}
	}()
	return httpListener, raftListener, nil
}

func initWebsocket(connectionReadLimit int64, listener net.Listener, raftStore *store.Store) {
	etpConfig := etp.ServerConfig{
		InsecureSkipVerify:  true,
		ConnectionReadLimit: connectionReadLimit,
	}
	etpServer := etp.NewServer(context.Background(), etpConfig)
	subs.NewSocketEventHandler(etpServer, raftStore).SubscribeAll()

	mux := http.NewServeMux()
	mux.HandleFunc("/isp-etp/", etpServer.ServeHttp)
	httpServer := &http.Server{Handler: mux}
	go func() {
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf(codes.StartHttpServerError, "http server closed: %v", err)
		}
	}()
	holder.EtpServer = etpServer
	holder.HTTPServer = httpServer
}

func initRaft(listener net.Listener, clusterCfg conf.ClusterConfiguration,
	declaration structure.BackendDeclaration) (*cluster.Client, *store.Store) {
	raftState, err := store.NewStateFromRepository()
	if err != nil {
		log.Fatal(codes.RestoreFromRepositoryError, err)
		return nil, nil
	}
	raftStore := store.NewStateStore(raftState)
	r, err := raft.NewRaft(listener, clusterCfg, raftStore)
	if err != nil {
		log.Fatalf(codes.InitRaftError, "unable to create raft server. %v", err)
		return nil, nil
	}
	clusterClient := cluster.NewRaftClusterClient(r, declaration, func(address string) {
		raftStore.VisitState(func(s state.WritableState) {
			cfg := config.Get().(*conf.Configuration)
			addr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				panic(err) // must never occurred
			}
			port := cfg.GrpcOuterAddress.Port
			addressConfiguration := structure.AddressConfiguration{Port: port, IP: addr.IP.String()}
			back := structure.BackendDeclaration{ModuleName: cfg.ModuleName, Address: addressConfiguration}
			log.WithMetadata(map[string]interface{}{"declaration": back}).
				Infof(codes.LeaderManualDeleteLeader, "manually delete disconnected leader's declaration")
			service.ClusterMesh.HandleDeleteBackendDeclarationCommand(back, s)
		})
	})
	holder.ClusterClient = clusterClient

	if clusterCfg.BootstrapCluster {
		err = r.BootstrapCluster() // err can be ignored
		if err != nil {
			log.Errorf(codes.BootstrapClusterError, "bootstrap cluster. %v", err)
		}
	}
	return clusterClient, raftStore
}

func initGrpc(bindAddress structure.AddressConfiguration, raftStore *store.Store) {
	controller.Routes = controller.NewRoutes(raftStore)
	controller.Schema = controller.NewSchema(raftStore)
	controller.Module = controller.NewModule(raftStore)
	controller.Config = controller.NewConfig(raftStore)
	controller.CommonConfig = controller.NewCommonConfig(raftStore)
	defaultService := backend.GetDefaultService(moduleName, helper.GetHandlers())
	backend.StartBackendGrpcServer(bindAddress, defaultService)
}

func gracefulShutdown() {
	const (
		gracefulTimeout  = 3 * time.Second
		terminateTimeout = 4 * time.Second
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	log.Infof(0, "received signal '%s'", <-quit)

	finishedCh := make(chan struct{})
	go func() {
		select {
		case <-time.After(terminateTimeout):
			log.Fatal(0, "exit timeout reached: terminating")
		case <-finishedCh:
			// ok
		case sig := <-quit:
			log.Fatalf(0, "received duplicate exit signal '%s': terminating", sig)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()
	onShutdown(ctx)
	finishedCh <- struct{}{}
	log.Info(0, "application exited normally")
}

func onShutdown(ctx context.Context) {
	backend.StopGrpcServer()
	holder.EtpServer.Close()

	if err := holder.HTTPServer.Shutdown(ctx); err != nil {
		log.Warnf(codes.ShutdownHttpServerError, "http server shutdown err: %v", err)
	} else {
		log.Info(codes.ShutdownHttpServerInfo, "http server shutdown success")
	}

	if err := holder.ClusterClient.Shutdown(); err != nil {
		log.Warnf(codes.RaftShutdownError, "raft shutdown err: %v", err)
	} else {
		log.Info(codes.RaftShutdownInfo, "raft shutdown success")
	}

	if err := model.DbClient.Close(); err != nil {
		log.Warnf(0, "database shutdown err: %v", err)
	}

	_ = muxer.Close()
}

func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "9.9.9.9:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil
}
