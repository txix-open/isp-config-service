package goose_store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3/database"
	"isp-config-service/repository"
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
	if _, err := s.db.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("failed to create version table %q: %w", s.tablename, err)
	}
	return nil
}

func (s Store) Insert(ctx context.Context, db database.DBTxConn, req database.InsertRequest) error {
	q := s.querier.InsertVersion(s.tablename)
	if _, err := s.db.ExecContext(ctx, q, req.Version, true); err != nil {
		return fmt.Errorf("failed to insert version %d: %w", req.Version, err)
	}
	return nil
}

func (s Store) Delete(ctx context.Context, db database.DBTxConn, version int64) error {
	q := s.querier.DeleteVersion(s.tablename)
	if _, err := s.db.ExecContext(ctx, q, version); err != nil {
		return fmt.Errorf("failed to delete version %d: %w", version, err)
	}
	return nil
}

func (s Store) GetMigration(ctx context.Context, db database.DBTxConn, version int64) (*database.GetMigrationResult, error) {
	q := s.querier.GetMigrationByVersion(s.tablename)
	var result database.GetMigrationResult

	isApplied := float64(0)
	err := s.db.QueryRowContext(ctx, q, version).Scan(&result.Timestamp, &isApplied)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %d", database.ErrVersionNotFound, version)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get migration %d: %w", version, err)
	}
	result.IsApplied = isApplied == 0
	return &result, nil
}

func (s Store) GetLatestVersion(ctx context.Context, db database.DBTxConn) (int64, error) {
	return -1, errors.New("not implemented")
}

func (s Store) ListMigrations(ctx context.Context, db database.DBTxConn) ([]*database.ListMigrationsResult, error) {
	q := s.querier.ListMigrations(s.tablename)
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %w", err)
	}
	defer rows.Close()

	var migrations []*database.ListMigrationsResult
	for rows.Next() {
		var result database.ListMigrationsResult
		isApplied := float64(0)
		version := float64(0)
		if err := rows.Scan(&version, &isApplied); err != nil {
			return nil, fmt.Errorf("failed to scan list migrations result: %w", err)
		}
		result.IsApplied = isApplied == 0
		result.Version = int64(version)
		migrations = append(migrations, &result)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithMessage(err, "fetch rows")
	}
	return migrations, nil
}
