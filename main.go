package main

import (
	"context"
	"github.com/integration-system/isp-etp-go"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"github.com/thecodeteam/goodbye"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/conf"
	"isp-config-service/controller"
	"isp-config-service/helper"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/raft"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"
	"isp-config-service/subs"
	"net"
	"net/http"
	"os"
	"strconv"
)

const (
	moduleName = "config"
)

var (
	version = "0.1.0"
	date    = "undefined"

	shutdownChan = make(chan struct{})
)

func init() {
	config.InitConfig(&conf.Configuration{})
}

// @title ISP configuration service
// @version 2.0.0
// @description Сервис управления конфигурацией модулей ISP кластера

// @license.name GNU GPL v3.0

// @host localhost:9003
// @BasePath /api/config
func main() {
	ctx := context.Background()
	cfg := config.Get().(*conf.Configuration)
	handlers := helper.GetHandlers()
	endpoints := backend.GetEndpoints(cfg.ModuleName, handlers)
	declaration := structure.BackendDeclaration{
		ModuleName: cfg.ModuleName,
		Version:    version,
		LibVersion: bootstrap.LibraryVersion,
		Endpoints:  endpoints,
		Address:    cfg.WS.Grpc,
	}

	model.DbClient.ReceiveConfiguration(cfg.Database)
	_, raftStore := initRaft(cfg.WS.Raft.GetAddress(), cfg.Cluster, declaration)
	initWebsocket(ctx, cfg.WS.Rest.GetAddress(), raftStore)
	initGrpc(cfg.WS.Grpc, raftStore)

	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(onShutdown)

	<-shutdownChan
}

func initWebsocket(ctx context.Context, bindAddress string, raftStore *store.Store) {
	etpConfig := etp.ServerConfig{
		InsecureSkipVerify: true,
	}
	etpServer := etp.NewServer(ctx, etpConfig)
	subs.NewSocketEventHandler(etpServer, raftStore).SubscribeAll()

	mux := http.NewServeMux()
	mux.HandleFunc("/isp-etp/", etpServer.ServeHttp)
	httpServer := &http.Server{Addr: bindAddress, Handler: mux}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf(codes.StartHttpServerError, "http server closed: %v", err)
		}
	}()
	holder.EtpServer = etpServer
	holder.HttpServer = httpServer
}

func initRaft(bindAddress string, clusterCfg conf.ClusterConfiguration, declaration structure.BackendDeclaration) (*cluster.Client, *store.Store) {
	raftState, err := store.NewStateFromRepository()
	if err != nil {
		log.Fatal(codes.RestoreFromRepositoryError, err)
		return nil, nil
	}
	raftStore := store.NewStateStore(raftState)
	r, err := raft.NewRaft(bindAddress, clusterCfg, raftStore)
	if err != nil {
		log.Fatalf(codes.InitRaftError, "unable to create raft server. %v", err)
		return nil, nil
	}
	clusterClient := cluster.NewRaftClusterClient(r, declaration, func(address string) {
		raftStore.VisitState(func(s state.WritableState) {
			cfg := config.Get().(*conf.Configuration)
			addr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				panic(err) // must never occured
			}
			port := addr.Port
			// TODO логика для определения порта пира, т.к всё тестируется на одной машине
			//peerNumber := addr.Port - 9000
			//switch peerNumber {
			//case 2:
			//	port = 9022
			//case 3:
			//	port = 9032
			//case 4:
			//	port = 9042
			//}
			//
			addressConfiguration := structure.AddressConfiguration{Port: strconv.Itoa(port), IP: addr.IP.String()}
			back := structure.BackendDeclaration{ModuleName: cfg.ModuleName, Address: addressConfiguration}
			service.ClusterMeshService.HandleDeleteBackendDeclarationCommand(back, s)
		})
	})
	holder.ClusterClient = clusterClient

	err = r.BootstrapCluster() // err can be ignored
	if err != nil {
		log.Errorf(codes.BootstrapClusterError, "bootstrap cluster. %v", err)
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

func onShutdown(ctx context.Context, sig os.Signal) {
	defer close(shutdownChan)

	backend.StopGrpcServer()

	if err := holder.ClusterClient.Shutdown(); err != nil {
		log.Warnf(codes.RaftShutdownError, "raft shutdown err: %v", err)
	} else {
		log.Info(codes.RaftShutdownInfo, "raft shutdown success")
	}

	holder.EtpServer.Close()

	if err := holder.HttpServer.Shutdown(ctx); err != nil {
		log.Warnf(codes.ShutdownHttpServerError, "http server shutdown err: %v", err)
	} else {
		log.Info(codes.ShutdownHttpServerInfo, "http server shutdown success")
	}
}
