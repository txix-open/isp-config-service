package assembly

import (
	"net/http"
	"time"

	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/worker"
	"isp-config-service/controller"
	"isp-config-service/controller/api"
	"isp-config-service/repository"
	"isp-config-service/routes"
	apisvs "isp-config-service/service/api"
	"isp-config-service/service/event"
	"isp-config-service/service/module"
	"isp-config-service/service/rqlite/db"
	"isp-config-service/service/subscription"
)

const (
	wsReadLimit          = 4 * 1024 * 1024
	handleEventsInterval = 500 * time.Millisecond
	cleanEventInterval   = 60 * time.Second
	eventTtl             = 60 * time.Second
)

type Locator struct {
	db     db.DB
	logger log.Logger
}

func NewLocator(logger log.Logger, db db.DB) Locator {
	return Locator{
		db:     db,
		logger: logger,
	}
}

type Config struct {
	GrpcMux           *grpc.Mux
	HttpMux           http.Handler
	EtpSrv            *etp.Server
	HandleEventWorker *worker.Worker
	CleanEventWorker  *worker.Worker
}

func (l Locator) Config() Config {
	moduleRepo := repository.NewModule(l.db)
	backendRepo := repository.NewBackend(l.db)
	eventRepo := repository.NewEvent(l.db)
	configRepo := repository.NewConfig(l.db)
	configSchemaRepo := repository.NewConfigSchema(l.db)
	configHistoryRepo := repository.NewConfigHistory(l.db)

	etpSrv := etp.NewServer(
		etp.WithServerReadLimit(wsReadLimit),
		etp.WithServerAcceptOptions(&etp.AcceptOptions{
			InsecureSkipVerify: true,
		}),
	)
	emitter := module.NewEmitter(l.logger)
	subscriptionService := subscription.NewService(
		moduleRepo,
		backendRepo,
		configRepo,
		etpSrv.Rooms(),
		emitter,
		l.logger,
	)

	moduleService := module.NewService(
		moduleRepo,
		backendRepo,
		eventRepo,
		configRepo,
		configSchemaRepo,
		subscriptionService,
		emitter,
		l.logger,
	)
	moduleController := controller.NewModule(moduleService, l.logger)

	moduleApiService := apisvs.NewModule(moduleRepo, backendRepo, configSchemaRepo)
	moduleApiController := api.NewModule(moduleApiService)

	configApiService := apisvs.NewConfig(
		configRepo,
		moduleRepo,
		configSchemaRepo,
		eventRepo,
		configHistoryRepo,
	)
	configApiController := api.NewConfig(configApiService)

	configHistoryApiService := apisvs.NewConfigHistory(configHistoryRepo)
	configHistoryController := api.NewConfigHistory(configHistoryApiService)

	configSchemaApiService := apisvs.NewConfigSchema(configSchemaRepo)
	configSchemaController := api.NewConfigSchema(configSchemaApiService)

	controllers := routes.Controllers{
		Module:           moduleController,
		ModuleApi:        moduleApiController,
		ConfigApi:        configApiController,
		ConfigHistoryApi: configHistoryController,
		ConfigSchemaApi:  configSchemaController,
	}
	mapper := endpoint.DefaultWrapper(l.logger)
	grpcMux := routes.GrpcHandler(mapper, controllers)

	routes.BindEtp(etpSrv, controllers, l.logger)

	httpMux := routes.HttpHandler(etpSrv)

	eventHandler := event.NewHandler(subscriptionService, l.logger)
	handleEventJob := event.NewWorker(eventRepo, eventHandler, l.logger)
	handleEventWorker := worker.New(handleEventJob, worker.WithInterval(handleEventsInterval))

	cleanerJob := event.NewCleaner(eventRepo, eventTtl, l.logger)
	cleanEventWorker := worker.New(cleanerJob, worker.WithInterval(cleanEventInterval))

	return Config{
		GrpcMux:           grpcMux,
		EtpSrv:            etpSrv,
		HttpMux:           httpMux,
		HandleEventWorker: handleEventWorker,
		CleanEventWorker:  cleanEventWorker,
	}
}
