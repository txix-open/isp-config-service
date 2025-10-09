package assembly

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"isp-config-service/conf"
	"isp-config-service/service/module/backend"
	"isp-config-service/service/variable"

	httpEndpoint "github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"

	"isp-config-service/controller"
	"isp-config-service/controller/api"
	"isp-config-service/repository"
	"isp-config-service/routes"
	apisvs "isp-config-service/service/api"
	"isp-config-service/service/event"
	"isp-config-service/service/metrics"
	"isp-config-service/service/module"
	"isp-config-service/service/rqlite/db"
	"isp-config-service/service/subscription"

	"github.com/txix-open/etp/v4"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/worker"
)

const (
	wsReadLimit                  = 4 * 1024 * 1024
	handleEventsInterval         = 500 * time.Millisecond
	cleanEventInterval           = 60 * time.Second
	cleanPhantomBackendsInterval = 5 * time.Minute
	eventTtl                     = 60 * time.Second
)

type LocalConfig struct {
	Local         conf.Local
	RqliteAddress string
}

type Locator struct {
	db            db.DB
	cfg           LocalConfig
	leaderChecker LeaderChecker
	logger        log.Logger
}

type LeaderChecker interface {
	IsLeader() bool
}

type OwnBackendsCleaner interface {
	DeleteOwnBackends(ctx context.Context)
}

func NewLocator(
	db db.DB,
	leaderChecker LeaderChecker,
	cfg LocalConfig,
	logger log.Logger,
) Locator {
	return Locator{
		db:            db,
		leaderChecker: leaderChecker,
		cfg:           cfg,
		logger:        logger,
	}
}

type Config struct {
	GrpcMux                   *grpc.Mux
	HttpMux                   http.Handler
	MetricsAdHandler          http.Handler
	EtpSrv                    *etp.Server
	HandleEventWorker         *worker.Worker
	CleanEventWorker          *worker.Worker
	CleanPhantomBackendWorker *worker.Worker
	OwnBackendsCleaner        OwnBackendsCleaner
}

//nolint:funlen
func (l Locator) Config() *Config {
	moduleRepo := repository.NewModule(l.db)
	backendRepo := repository.NewBackend(l.db)
	eventRepo := repository.NewEvent(l.db)
	configRepo := repository.NewConfig(l.db)
	configSchemaRepo := repository.NewConfigSchema(l.db)
	configHistoryRepo := repository.NewConfigHistory(l.db)
	variableRepo := repository.NewVariable(l.db)

	variableService := variable.NewService(variableRepo, configRepo, eventRepo)
	variableController := api.NewVariable(variableService)

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
		variableService,
		l.logger,
	)

	backendService := backend.NewBackend(
		backendRepo,
		eventRepo,
		l.cfg.Local.Rqlite.NodeId,
		etpSrv.Rooms(),
		l.logger,
	)
	cleanPhantomBackendJob := backend.NewCleaner(backendService, l.logger)
	cleanPhantomBackendWorker := worker.New(cleanPhantomBackendJob, worker.WithInterval(cleanPhantomBackendsInterval))

	moduleService := module.NewService(
		moduleRepo,
		backendService,
		configRepo,
		configSchemaRepo,
		subscriptionService,
		emitter,
		variableService,
		l.logger,
	)
	moduleController := controller.NewModule(moduleService, l.logger)

	configHistoryApiService := apisvs.NewConfigHistory(configRepo, moduleRepo, configHistoryRepo,
		l.cfg.Local.KeepConfigVersions, l.logger)
	configHistoryController := api.NewConfigHistory(configHistoryApiService)

	moduleApiService := apisvs.NewModule(moduleRepo, backendRepo, configSchemaRepo)
	moduleApiController := api.NewModule(moduleApiService)

	configApiService := apisvs.NewConfig(
		configRepo,
		moduleRepo,
		configSchemaRepo,
		eventRepo,
		configHistoryApiService,
		variableService,
		l.logger,
	)
	configApiController := api.NewConfig(configApiService)

	configSchemaApiService := apisvs.NewConfigSchema(configSchemaRepo)
	configSchemaController := api.NewConfigSchema(configSchemaApiService)

	controllers := routes.Controllers{
		Module:           moduleController,
		ModuleApi:        moduleApiController,
		ConfigApi:        configApiController,
		ConfigHistoryApi: configHistoryController,
		ConfigSchemaApi:  configSchemaController,
		VariableApi:      variableController,
	}
	mapper := endpoint.DefaultWrapper(l.logger)
	grpcMux := routes.GrpcHandler(mapper, controllers)

	routes.BindEtp(etpSrv, controllers, l.logger)

	rqliteUrl, err := url.Parse(l.cfg.RqliteAddress)
	if err != nil {
		panic(err)
	}
	rqliteProxy := httputil.NewSingleHostReverseProxy(rqliteUrl)
	httpMux := routes.HttpHandler(etpSrv, l.cfg.Local, rqliteProxy)

	eventHandler := event.NewHandler(subscriptionService, l.logger)
	handleEventJob := event.NewWorker(eventRepo, eventHandler, l.logger)
	handleEventWorker := worker.New(handleEventJob, worker.WithInterval(handleEventsInterval))

	cleanerJob := event.NewCleaner(eventRepo, l.leaderChecker, eventTtl, l.logger)
	cleanEventWorker := worker.New(cleanerJob, worker.WithInterval(cleanEventInterval))

	metricsSvc := metrics.New(backendRepo)
	metricsController := controller.NewMetrics(metricsSvc)
	httpWrapper := httpEndpoint.DefaultWrapper(l.logger, httplog.Log(l.logger, false))

	return &Config{
		GrpcMux:                   grpcMux,
		EtpSrv:                    etpSrv,
		HttpMux:                   httpMux,
		MetricsAdHandler:          httpWrapper.Endpoint(metricsController.Autodiscovery),
		HandleEventWorker:         handleEventWorker,
		CleanEventWorker:          cleanEventWorker,
		CleanPhantomBackendWorker: cleanPhantomBackendWorker,
		OwnBackendsCleaner:        backendService,
	}
}
