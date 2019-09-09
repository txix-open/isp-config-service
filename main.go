package main

import (
	"context"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/thecodeteam/goodbye"
	"isp-config-service/cluster"
	"isp-config-service/conf"
	"isp-config-service/helper"
	"isp-config-service/holder"
	"isp-config-service/raft"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"
	"isp-config-service/subs"
	"isp-config-service/ws"
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

func main() {
	cfg := config.Get().(*conf.Configuration)
	handlers := helper.GetAllHandlers()
	endpoints := backend.GetEndpoints(cfg.ModuleName, handlers...)
	declaration := structure.BackendDeclaration{
		ModuleName: cfg.ModuleName,
		Version:    version,
		LibVersion: bootstrap.LibraryVersion,
		Endpoints:  endpoints,
		Address:    cfg.WS.Grpc,
	}

	client, store := initRaft(cfg.WS.Raft.GetAddress(), cfg.Cluster, declaration)
	initWebsocket(cfg.WS.Rest.GetAddress(), client, store)
	initGrpc(cfg.WS.Grpc)

	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(onShutdown)

	<-shutdownChan
}

func initWebsocket(bindAddress string, clusterClient *cluster.ClusterClient, store *store.Store) {
	socket, err := ws.NewWebsocketServer()
	if err != nil {
		logger.Fatal(err)
	}
	subs.NewSocketEventHandler(socket, clusterClient, store).SubscribeAll()

	mux := http.NewServeMux()
	mux.HandleFunc("/socket.io/", socket.ServeHttp)
	httpServer := &http.Server{Addr: bindAddress, Handler: mux}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()
	holder.Socket = socket
	holder.HttpServer = httpServer

	logger.Infof("socket.IO server start on %s", bindAddress)
}

func initRaft(bindAddress string, clusterCfg conf.ClusterConfiguration, declaration structure.BackendDeclaration) (*cluster.ClusterClient, *store.Store) {
	store := store.NewStateStore()
	r, err := raft.NewRaft(bindAddress, clusterCfg, store)
	if err != nil {
		logger.Fatal(err)
	}
	clusterClient := cluster.NewRaftClusterClient(r, declaration, func(address string) {
		store.VisitState(func(s state.State) {
			cfg := config.Get().(*conf.Configuration)
			addr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				panic(err) //must never occured
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
			_, err = service.ClusterStateService.HandleDeleteBackendDeclarationCommand(back, s)
			if err != nil {
				logger.Debug("onClientDisconnect HandleDeleteBackendDeclarationCommand", address, err)
			}
		})
	})
	holder.ClusterClient = clusterClient

	err = r.BootstrapCluster() // err can be ignored
	if err != nil {
		logger.Error("BootstrapCluster err:", err)
	}

	logger.Infof("raft server start on %s", bindAddress)

	return clusterClient, store
}

func initGrpc(bindAddress structure.AddressConfiguration) {
	service := backend.GetDefaultService(moduleName, helper.GetAllHandlers()...)
	backend.StartBackendGrpcServer(bindAddress, service)
}

func onShutdown(ctx context.Context, sig os.Signal) {
	defer close(shutdownChan)

	backend.StopGrpcServer()

	if err := holder.ClusterClient.Shutdown(); err != nil {
		logger.Warnf("raft shutdown err: %v", err)
	} else {
		logger.Info("raft shutdown success")
	}

	if err := holder.Socket.Close(); err != nil {
		logger.Warnf("socket.io shutdown err: %v", err)
	} else {
		logger.Info("socket.io shutdown success")
	}

	if err := holder.HttpServer.Shutdown(ctx); err != nil {
		logger.Warnf("http server shutdown err: %v", err)
	} else {
		logger.Info("http server shutdown success")
	}
}
