package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/dbx/migration"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/assembly"
	"isp-config-service/conf"
	"isp-config-service/service/rqlite"
	"isp-config-service/service/rqlite/db"
	"isp-config-service/service/rqlite/goose_store"
)

const (
	waitForLeaderTimeout = 30 * time.Second
)

type Service struct {
	boot    *bootstrap.Bootstrap
	rqlite  *rqlite.Rqlite
	grpcSrv *grpc.Server
	httpSrv *http.Server
	logger  log.Logger

	//initialized in Run
	etpSrv *etp.Server
}

func New(boot *bootstrap.Bootstrap) Service {
	rqlite := rqlite.New(boot.App.Config())
	return Service{
		boot:    boot,
		rqlite:  rqlite,
		logger:  boot.App.Logger(),
		httpSrv: http.NewServer(boot.App.Logger()),
		grpcSrv: grpc.DefaultServer(),
	}
}

func (s Service) Run(ctx context.Context) error {
	localConfig := conf.Local{}
	err := s.boot.App.Config().Read(&localConfig)
	if err != nil {
		return errors.WithMessage(err, "read local config")
	}

	go func() {
		s.logger.Debug(ctx, "running embedded rqlite...")
		err := s.rqlite.Run(ctx)
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "run embedded rqlite"))
		}
	}()
	time.Sleep(1 * time.Second) //optimistically wait for store initialization

	s.logger.Debug(ctx, fmt.Sprintf("waiting for cluster startup for %s...", waitForLeaderTimeout))
	err = s.rqlite.WaitForLeader(waitForLeaderTimeout)
	if err != nil {
		return errors.WithMessage(err, "wait for leader")
	}

	if s.rqlite.IsLeader() {
		s.logger.Debug(ctx, "is a leader")
		err := s.leaderStartup(ctx)
		if err != nil {
			return errors.WithMessage(err, "start leader")
		}
	} else {
		s.logger.Debug(ctx, "is not a leader")
	}

	db, err := db.Open(ctx, s.rqlite.Dsn(), httpclix.Default())
	if err != nil {
		return errors.WithMessage(err, "dial to embedded rqlite")
	}

	config := assembly.NewLocator(s.logger, db).Config()
	s.etpSrv = config.EtpSrv
	s.grpcSrv.Upgrade(config.GrpcMux)
	s.httpSrv.Upgrade(config.HttpMux)

	go func() {
		s.logger.Debug(ctx, fmt.Sprintf("starting grpc server on %s", s.boot.BindingAddress))
		err := s.grpcSrv.ListenAndServe(s.boot.BindingAddress)
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "start grpc server"))
		}
	}()

	go func() {
		httpPort := localConfig.ConfigServiceAddress.Port
		s.logger.Debug(ctx, fmt.Sprintf("starting http server on 0.0.0.0:%s", httpPort))
		err := s.httpSrv.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", httpPort))
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "start http server"))
		}
	}()
	time.Sleep(1 * time.Second)

	return nil
}

func (s Service) Closers() []app.Closer {
	return []app.Closer{
		app.CloserFunc(func() error {
			s.grpcSrv.Shutdown()
			return nil
		}),
		app.CloserFunc(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return s.httpSrv.Shutdown(ctx)
		}),
		app.CloserFunc(func() error {
			if s.etpSrv == nil {
				return nil
			}
			s.etpSrv.OnDisconnect(nil)
			s.etpSrv.Shutdown()
			return nil
		}),
		s.rqlite,
	}
}

func (s Service) leaderStartup(ctx context.Context) error {
	db, err := s.rqlite.SqlDB()
	if err != nil {
		return errors.WithMessage(err, "open sql db")
	}
	defer db.Close()

	migrationRunner := migration.NewRunner("", s.boot.MigrationsDir, s.logger)
	rqliteGooseStore := goose_store.NewStore(db)
	err = migrationRunner.Run(ctx, db, goose.WithStore(rqliteGooseStore))
	if err != nil {
		return errors.WithMessage(err, "apply migrations")
	}

	return nil
}
