package migration

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/txix-open/isp-kit/log"
)

const (
	DialectPostgreSQL = goose.DialectPostgres
	DialectSqlite3    = goose.DialectSQLite3
	DialectClickhouse = goose.DialectClickHouse
)

type Runner struct {
	dialect goose.Dialect
	fs      fs.FS
	logger  log.Logger
}

func NewRunner(
	dialect goose.Dialect,
	fs fs.FS,
	logger log.Logger,
) Runner {
	return Runner{
		dialect: dialect,
		fs:      fs,
		logger:  logger,
	}
}

func (r Runner) Run(ctx context.Context, db *sql.DB, gooseOpts ...goose.ProviderOption) error {
	ctx = log.ToContext(ctx, log.String("worker", "goose_db_migration"))

	provider, err := goose.NewProvider(r.dialect, db, r.fs, gooseOpts...)
	if err != nil {
		return errors.WithMessage(err, "get goose provider")
	}

	dbVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return errors.WithMessage(err, "get db version")
	}
	r.logger.Info(ctx, fmt.Sprintf("current db version: %d", dbVersion))

	migrations, err := provider.Status(ctx)
	if err != nil {
		return errors.WithMessage(err, "get status migrations")
	}
	r.logger.Info(ctx, "print migration list")
	if len(migrations) == 0 {
		r.logger.Info(ctx, "no migrations")
	}
	for _, migration := range migrations {
		appliedAt := "Pending"
		if !migration.AppliedAt.IsZero() {
			appliedAt = migration.AppliedAt.Format(time.RFC3339)
		}
		msg := fmt.Sprintf(
			"migration: %s %s %s",
			filepath.Base(migration.Source.Path),
			strings.ToUpper(string(migration.State)),
			appliedAt,
		)
		r.logger.Info(ctx, msg)
	}

	result, err := provider.Up(ctx)
	if err != nil {
		return errors.WithMessage(err, "apply pending migrations")
	}
	for _, migrationResult := range result {
		r.logger.Info(ctx, fmt.Sprintf("applied migration: %s", migrationResult.String()))
	}

	return nil
}
