package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/dbx/migration"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/assembly"
	"isp-config-service/service/rqlite"
	"isp-config-service/service/rqlite/db"
	"isp-config-service/service/rqlite/goose_store"
)

const (
	waitForLeaderTimeout = 30 * time.Second
)

type Service struct {
	boot   *bootstrap.Bootstrap
	logger log.Logger
	rqlite *rqlite.Rqlite
}

func New(boot *bootstrap.Bootstrap) Service {
	rqlite := rqlite.New(boot.App.Config())
	return Service{
		boot:   boot,
		logger: boot.App.Logger(),
		rqlite: rqlite,
	}
}

func (s Service) Run(ctx context.Context) error {
	go func() {
		err := s.rqlite.Run(ctx)
		if err != nil {
			s.boot.Fatal(errors.WithMessage(err, "run embedded rqlite"))
		}
	}()
	time.Sleep(1 * time.Second) //optimistically wait for store initialization

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

	db, err := db.Open(ctx, s.rqlite.Dsn(), httpclix.Default())
	if err != nil {
		return errors.WithMessage(err, "dial to embedded rqlite")
	}

	_ = assembly.NewLocator(s.logger, db)

	return nil
}

func (s Service) Close() error {
	_ = s.rqlite.Close()
	return nil
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
