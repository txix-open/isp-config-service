package db

import (
	"context"
	"database/sql"
)

type DB interface {
	Select(ctx context.Context, ptr any, query string, args ...any) error
	SelectRow(ctx context.Context, ptr any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error)
	ExecTransaction(ctx context.Context, requests [][]any) ([]sql.Result, error)
}
