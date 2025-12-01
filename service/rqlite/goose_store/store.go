package goose_store

import (
	"context"
	"database/sql"

	"isp-config-service/repository"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3/database"
)

type Store struct {
	tablename string
	querier   *Rqlite
	db        *sql.DB
}

func NewStore(db *sql.DB) Store {
	return Store{
		db:        db,
		tablename: repository.Table("goose_db_version"),
		querier:   &Rqlite{},
	}
}

func (s Store) Tablename() string {
	return s.tablename
}

func (s Store) CreateVersionTable(ctx context.Context, db database.DBTxConn) error {
	q := s.querier.CreateTable(s.tablename)
	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return errors.WithMessagef(err, "failed to create version table %q", s.tablename)
	}
	return nil
}

func (s Store) Insert(ctx context.Context, db database.DBTxConn, req database.InsertRequest) error {
	q := s.querier.InsertVersion(s.tablename)
	_, err := s.db.ExecContext(ctx, q, req.Version, true)
	if err != nil {
		return errors.WithMessagef(err, "failed to insert version %d", req.Version)
	}
	return nil
}

func (s Store) Delete(ctx context.Context, db database.DBTxConn, version int64) error {
	q := s.querier.DeleteVersion(s.tablename)
	_, err := s.db.ExecContext(ctx, q, version)
	if err != nil {
		return errors.WithMessagef(err, "failed to delete version %d", version)
	}
	return nil
}

func (s Store) GetMigration(ctx context.Context, db database.DBTxConn, version int64) (*database.GetMigrationResult, error) {
	q := s.querier.GetMigrationByVersion(s.tablename)
	var result database.GetMigrationResult

	isApplied := float64(0)
	err := s.db.QueryRowContext(ctx, q, version).Scan(&result.Timestamp, &isApplied)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.WithMessagef(database.ErrVersionNotFound, "%d", version)
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get migration %d", version)
	}
	result.IsApplied = isApplied == 0
	return &result, nil
}

func (s Store) GetLatestVersion(ctx context.Context, db database.DBTxConn) (int64, error) {
	return -1, nil
}

func (s Store) ListMigrations(ctx context.Context, db database.DBTxConn) ([]*database.ListMigrationsResult, error) {
	q := s.querier.ListMigrations(s.tablename)
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to list migrations")
	}
	defer rows.Close()

	var migrations []*database.ListMigrationsResult
	for rows.Next() {
		var result database.ListMigrationsResult
		isApplied := float64(0)
		version := float64(0)
		err = rows.Scan(&version, &isApplied)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to scan list migrations result")
		}
		result.IsApplied = isApplied == 0
		result.Version = int64(version)
		migrations = append(migrations, &result)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.WithMessage(err, "fetch rows")
	}
	return migrations, nil
}
