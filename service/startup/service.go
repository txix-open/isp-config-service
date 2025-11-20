package startup

import (
	"context"
	"fmt"
	"time"

	"isp-config-service/assembly"
	"isp-config-service/conf"
	"isp-config-service/middlewares"
	"isp-config-service/service/rqlite"
	"isp-config-service/service/rqlite/db"
	"isp-config-service/service/rqlite/goose_store"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/dbx/migration"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/observability/sentry"
)

const (
	waitForLeaderTimeout = 30 * time.Second
)

type Service struct {
	boot       *bootstrap.Bootstrap
	cfg        conf.Local
	rqlite     *rqlite.Rqlite
	grpcSrv    *grpc.Server
	httpSrv    *http.Server
	clusterCli bootstrap.ClusterClient
	logger     log.Logger

	// initialized in Run
	locatorConfig *assembly.Config
}

func New(boot *bootstrap.Bootstrap) (*Service, error) {
	localConfig := conf.Local{}
	err := boot.App.Config().Read(&localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "read local config")
	}

	logLevelValue := boot.App.Config().Optional().String("logLevel", "info")
	var logLevel log.Level
	err = logLevel.UnmarshalText([]byte(logLevelValue))
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal log level")
	}
	boot.App.Logger().SetLevel(logLevel)

	internalClientCredential, err := internalClientCredentials(localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "get internal client credentials")
	}

	rqlite := rqlite.New(boot.App.Config(), internalClientCredential, localConfig.Credentials)

	return &Service{
		boot:       boot,
		cfg:        localConfig,
		rqlite:     rqlite,
		grpcSrv:    grpc.DefaultServer(),
		httpSrv:    http.NewServer(boot.App.Logger()),
		clusterCli: boot.ClusterCli,
		logger:     sentry.WrapErrorLogger(boot.App.Logger(), boot.SentryHub),
	}, nil
}

// nolint:funlen
func (s *Service) Run(ctx context.Context) error {
	if s.boot.App.Config().Optional().String("KUBERNETES_SERVICE_HOST", "") != "" {
		s.logger.Info(
			ctx,
			"run in kubernetes, sleep extra 5s, reason described here: https://github.com/rqlite/rqlite/blob/master/docker-entrypoint.sh#L96",
		)
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second): //nolint:mnd
		}
	}

	go func() {
		s.logger.Debug(ctx, "running embedded rqlite...")
		err := s.rqlite.Run(ctx)
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "run embedded rqlite"))
		}
	}()
	time.Sleep(1 * time.Second) // optimistically wait for store initialization

	s.logger.Debug(ctx, fmt.Sprintf("waiting for cluster startup for %s...", waitForLeaderTimeout))
	err := s.rqlite.WaitForLeader(waitForLeaderTimeout)
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
	time.Sleep(1 * time.Second)

	rqliteClient := httpclix.Default(httpcli.WithMiddlewares(middlewares.SqlOperationMiddleware()))
	rqliteClient.GlobalRequestConfig().BaseUrl = s.rqlite.LocalHttpAddr()
	rqliteClient.GlobalRequestConfig().BasicAuth = s.rqlite.InternalClientCredential()
	db, err := db.Open(
		ctx,
		rqliteClient,
	)
	if err != nil {
		return errors.WithMessage(err, "dial to embedded rqlite")
	}

	cfg := assembly.LocalConfig{
		Local:         s.cfg,
		RqliteAddress: s.rqlite.LocalHttpAddr(),
	}
	s.locatorConfig = assembly.NewLocator(db, s.rqlite, cfg, s.logger).Config()
	s.grpcSrv.Upgrade(s.locatorConfig.GrpcMux)
	s.httpSrv.Upgrade(s.locatorConfig.HttpMux)
	s.boot.InfraServer.Handle("/internal/metrics/autodiscovery", s.locatorConfig.MetricsAdHandler)

	s.locatorConfig.HandleEventWorker.Run(ctx)
	s.locatorConfig.CleanEventWorker.Run(ctx)
	s.locatorConfig.CleanPhantomBackendWorker.Run(ctx)

	if cfg.Local.Backup.Enabled {
		s.locatorConfig.CleanBackupsWorker.Run(ctx)
	}

	go func() {
		s.logger.Debug(ctx, fmt.Sprintf("starting grpc server on %s", s.boot.BindingAddress))
		err := s.grpcSrv.ListenAndServe(s.boot.BindingAddress)
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "start grpc server"))
		}
	}()

	go func() {
		httpPort := s.cfg.ConfigServiceAddress.Port
		s.logger.Debug(ctx, fmt.Sprintf("starting http server on 0.0.0.0:%s", httpPort))
		err := s.httpSrv.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", httpPort))
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "start http server"))
		}
	}()
	time.Sleep(1 * time.Second) // wait for http start

	if !s.cfg.MaintenanceMode {
		go func() {
			err = s.clusterCli.Run(ctx, cluster.NewEventHandler())
			if err != nil {
				s.boot.Fatal(errors.WithMessage(err, "connect to it's self"))
			}
		}()
	}

	return nil
}

func (s *Service) Closers() []app.Closer {
	return []app.Closer{
		s.clusterCli,
		app.CloserFunc(func() error {
			s.grpcSrv.Shutdown()
			return nil
		}),
		app.CloserFunc(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			return s.httpSrv.Shutdown(ctx)
		}),
		app.CloserFunc(func() error {
			if s.locatorConfig == nil {
				return nil
			}

			s.locatorConfig.HandleEventWorker.Shutdown()
			s.locatorConfig.CleanEventWorker.Shutdown()
			s.locatorConfig.CleanPhantomBackendWorker.Shutdown()

			s.locatorConfig.EtpSrv.OnDisconnect(nil)
			s.locatorConfig.EtpSrv.Shutdown()

			s.locatorConfig.OwnBackendsCleaner.DeleteOwnBackends(context.Background())

			return nil
		}),
		s.rqlite,
	}
}

func (s *Service) leaderStartup(ctx context.Context) error {
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

func internalClientCredentials(cfg conf.Local) (*httpcli.BasicAuth, error) {
	for _, credential := range cfg.Credentials {
		if credential.Username == cfg.InternalClientCredential {
			return &httpcli.BasicAuth{Username: credential.Username, Password: credential.Password}, nil
		}
	}
	return nil, errors.Errorf("internal client credential '%s' not found", cfg.InternalClientCredential)
}
