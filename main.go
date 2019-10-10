package main

import (
	"context"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"github.com/thecodeteam/goodbye"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/conf"
	"isp-config-service/helper"
	"isp-config-service/holder"
	"isp-config-service/model"
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

	model.DbClient.ReceiveConfiguration(cfg.Database)
	client, raftStore := initRaft(cfg.WS.Raft.GetAddress(), cfg.Cluster, declaration)
	initWebsocket(cfg.WS.Rest.GetAddress(), client, raftStore)
	initGrpc(cfg.WS.Grpc)

	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(onShutdown)

	<-shutdownChan
}

func initWebsocket(bindAddress string, clusterClient *cluster.ClusterClient, raftStore *store.Store) {
	socket, err := ws.NewWebsocketServer()
	if err != nil {
		// TODO уйдет при миграции
		log.Fatalf(0, "init socket.io %s", err.Error())
	}
	subs.NewSocketEventHandler(socket, clusterClient, raftStore).SubscribeAll()

	mux := http.NewServeMux()
	mux.HandleFunc("/socket.io/", socket.ServeHttp)
	httpServer := &http.Server{Addr: bindAddress, Handler: mux}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf(codes.StartHttpServerError, "unable to start http server. %v", err)
		}
	}()
	holder.Socket = socket
	holder.HttpServer = httpServer
}

func initRaft(bindAddress string, clusterCfg conf.ClusterConfiguration, declaration structure.BackendDeclaration) (*cluster.ClusterClient, *store.Store) {
	raftStore := store.NewStateStoreFromRepository()
	r, err := raft.NewRaft(bindAddress, clusterCfg, raftStore)
	if err != nil {
		log.Fatalf(codes.InitRaftError, "unable to create raft server. %v", err)
	}
	clusterClient := cluster.NewRaftClusterClient(r, declaration, func(address string) {
		raftStore.VisitState(func(s state.State) {
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
				log.Warnf(codes.DeleteBackendDeclarationError, "onClientDisconnect delete backend declaration. %v", err)
			}
		})
	})
	holder.ClusterClient = clusterClient

	err = r.BootstrapCluster() // err can be ignored
	if err != nil {
		log.Errorf(codes.BootstrapClusterError, "bootstrap cluster. %v", err)
	}
	return clusterClient, raftStore
}

func initGrpc(bindAddress structure.AddressConfiguration) {
	defaultService := backend.GetDefaultService(moduleName, helper.GetAllHandlers()...)
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

	if err := holder.Socket.Close(); err != nil {
		// TODO уйдет при миграции
		log.Warnf(0, "socket.io shutdown err: %s", err.Error())
	} else {
		log.Info(0, "socket.io shutdown success")
	}

	if err := holder.HttpServer.Shutdown(ctx); err != nil {
		log.Warnf(codes.ShutdownHttpServerError, "http server shutdown err: %v", err)
	} else {
		log.Info(codes.ShutdownHttpServerInfo, "http server shutdown success")
	}
}
