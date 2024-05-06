package assembly

import (
	"net/http"

	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/controller"
	"isp-config-service/repository"
	"isp-config-service/routes"
	"isp-config-service/service"
)

const (
	wsReadLimit = 4 * 1024 * 1024
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
	GrpcMux *grpc.Mux
	HttpMux http.Handler
	EtpSrv  *etp.Server
}

func (l Locator) Config() Config {
	moduleRepo := repository.NewModule(l.db)
	backendRepo := repository.NewBackend(l.db)
	eventRepo := repository.NewEvent(l.db)
	configRepo := repository.NewConfig(l.db)
	configSchemaRepo := repository.NewConfigSchema(l.db)
	moduleService := service.NewModule(
		moduleRepo,
		backendRepo,
		eventRepo,
		configRepo,
		configSchemaRepo,
		l.logger,
	)
	moduleController := controller.NewModule(moduleService, l.logger)
	controllers := routes.Controllers{
		Module: moduleController,
	}
	mapper := endpoint.DefaultWrapper(l.logger)
	grpcMux := routes.GrpcHandler(mapper, controllers)

	etpSrv := etp.NewServer(
		etp.WithServerReadLimit(wsReadLimit),
		etp.WithServerAcceptOptions(&etp.AcceptOptions{
			InsecureSkipVerify: true,
		}),
	)
	routes.BindEtp(etpSrv, controllers)

	httpMux := routes.HttpHandler(etpSrv)

	return Config{
		GrpcMux: grpcMux,
		EtpSrv:  etpSrv,
		HttpMux: httpMux,
	}
}
