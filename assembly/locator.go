package assembly

import (
	"net/http"

	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/routes"
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
	controllers := routes.Controllers{}
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
