package main

import (
	"flag"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/valyala/fasthttp"
	"isp-config-service/domain"
	"isp-config-service/module"
	"net/http"
	"os"
	"os/signal"
	"time"

	"isp-config-service/conf"
	"isp-config-service/generated"
	"isp-config-service/helper"
	"isp-config-service/model"
	"isp-config-service/socket"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/integration-system/isp-lib/backend"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/metric"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	routerModule    = "router"
	converterModule = "converter"
)

var (
	configData   *conf.Configuration
	echoEndpoint *string
	version      = "0.1.0"
	date         = "undefined"
)

func init() {
	configData = config.InitConfig(&conf.Configuration{}).(*conf.Configuration)
	database.InitDbWithSchema(configData.Database, configData.Database.Schema)
	db := database.GetDBManager()
	model.InitDbManager(db.Db)
	model.SetSchema(configData.Database.Schema)
	echoEndpoint = flag.String("echo_endpoint", configData.WS.Grpc.GetAddress(), "endpoint of YourService")
	module.Version = version
}

// @title ISP configuration service
// @version 1.2.0
// @description Сервис управления конфигурацией модулей ISP кластера

// @license.name GNU GPL v3.0

// @host localhost:9003
// @BasePath /api/config
func main() {
	metric.InitCollectors(configData.Metrics, structure.MetricConfiguration{})
	metric.InitStatusChecker("routers", checkRouters)
	metric.InitStatusChecker("converters", checkConverters)
	metric.InitHttpServer(configData.Metrics)
	metric.GerHttpRouter().GET("/modules", handleModulesRequest)

	createGrpcServer()
	createRestServer(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
}

func checkRouters() interface{} {
	result := make([]string, 0)

	instances, err := model.InstanceRep.GetInstances(nil)
	if err != nil {
		logger.Warn(err)
	}
	for _, v := range instances {
		list := socket.GetModuleAddressList(v.Uuid, routerModule)
		for _, v := range list {
			result = append(result, v.GetAddress())
		}
	}

	return result
}

func checkConverters() interface{} {
	result := make([]string, 0)
	instances, err := model.InstanceRep.GetInstances(nil)
	if err != nil {
		logger.Warn(err)
	}
	for _, v := range instances {
		list := socket.GetModuleAddressList(v.Uuid, converterModule)
		for _, v := range list {
			result = append(result, v.GetAddress())
		}
	}
	return result
}

// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(ctx context.Context, opts ...runtime.ServeMuxOption) (http.Handler, error) {
	mux := runtime.NewServeMux(opts...)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := isp.RegisterBackendServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, dialOpts)
	if err != nil {
		logger.Fatalf("failed to serve rest proxy: %v", err)
	}
	return mux, nil
}

// Start a GRPC server.
func createGrpcServer() {
	// Run our server in a goroutine so that it doesn't block.
	handlers := helper.GetHandlers()
	addr := structure.AddressConfiguration{IP: configData.WS.Grpc.IP, Port: configData.WS.Grpc.Port}
	addrOuter := structure.AddressConfiguration{IP: configData.GrpcOuterAddress.IP, Port: configData.GrpcOuterAddress.Port}
	backend.StartBackendGrpcServer(addr, backend.GetDefaultService(configData.ModuleName, handlers))
	enndpoints := backend.GetEndpoints(configData.ModuleName, handlers)

	socket.SetBackendMethods(&structure.BackendDeclaration{
		ModuleName: configData.ModuleName,
		Version:    version,
		LibVersion: bootstrap.LibraryVersion,
		Endpoints:  enndpoints,
		Address:    addrOuter,
	})
	logger.Infof("EXPORTED MODULE METHODS: %v, module_name: %s, version: %s",
		enndpoints, configData.ModuleName, version)
}

// Start a HTTP server.
func createRestServer(opts ...runtime.ServeMuxOption) {

	mux := http.NewServeMux()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// === Socket.IO ===
	mux.Handle("/socket.io/", socket.Get())

	// === REST ===
	/*mux.Handle("/swagger/", controller.StaticHandler(
	http.StripPrefix(
		"/swagger/", http.FileServer(
			http.Dir(path.Join(executableFileDir, "static", "swagger-ui"))))))*/

	gw, err := newGateway(ctx, opts...)
	if err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
	mux.Handle("/", gw)

	/*r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})*/

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15,
		"the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	restAddress := configData.WS.Rest.GetAddress()
	logger.Infof("Serving at %s ...", restAddress)
	srv := &http.Server{
		Addr: restAddress,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mux, // Pass our instance of gorilla/mux in.
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel = context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logger.Info("shutting down")
	os.Exit(0)
}

func handleModulesRequest(ctx *fasthttp.RequestCtx) {
	instances, err := model.InstanceRep.GetInstances(nil)
	if err != nil {
		logger.Warn(err)
		ctx.Error("internal server error", http.StatusInternalServerError)
		return
	}

	result := make(map[string][]*domain.ModuleInfo, len(instances))
	for _, i := range instances {
		info, err := module.GetAggregatedModuleInfo(i.Uuid)
		if err != nil {
			logger.Warn(err)
			continue
		}

		for _, inf := range info {
			inf.ConfigSchema = nil
			inf.Configs = nil
			for i := range inf.Status {
				s := &inf.Status[i]
				s.Endpoints = nil
			}
		}

		result[i.Uuid] = info
	}

	if bytes, err := utils.ConvertGoToBytes(result); err != nil {
		logger.Warn(err)
		ctx.Error("internal server error", http.StatusInternalServerError)
	} else {
		ctx.SetStatusCode(http.StatusOK)
		ctx.SetContentType("application/json; charset=utf-8")
		ctx.SetBody(bytes)
	}
}
