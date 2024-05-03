package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
)

type DB struct {
	cli *httpcli.Client
}

func Open(ctx context.Context, dsn string, client *httpcli.Client) (*DB, error) {
	client.GlobalRequestConfig().BaseUrl = dsn
	db := &DB{
		cli: client,
	}

	m := map[string]any{}
	err := db.SelectRow(ctx, &m, `SELECT 1 as test`)
	if err != nil {
		return nil, errors.WithMessage(err, "ping db")
	}

	return db, nil
}

func (d DB) Select(ctx context.Context, ptr any, query string, args ...any) error {
	result := &Result{
		Rows: ptr,
	}
	resp := Response{
		Results: []*Result{result},
	}
	err := d.cli.Post("/db/query").
		QueryParams(map[string]any{
			"timings":     true,
			"associative": true,
		}).JsonRequestBody(Request(query, args...)).
		JsonResponseBody(&resp).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return errors.WithMessage(err, "call rqlite")
	}
	if result.Error != "" {
		return errors.Errorf("sqlite: %s", result.Error)
	}
	return nil
}

func (d DB) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	result := &Result{}
	resp := Response{
		Results: []*Result{result},
	}
	httpResp, err := d.cli.Post("/db/query").
		QueryParams(map[string]any{
			"timings":     true,
			"associative": true,
		}).JsonRequestBody(Request(query, args...)).
		JsonResponseBody(&resp).
		StatusCodeToError().
		Do(ctx)
	if err != nil {
		return errors.WithMessage(err, "call rqlite")
	}
	defer httpResp.Close()

	if result.Error != "" {
		return errors.Errorf("sqlite: %s", result.Error)
	}

	body, _ := httpResp.Body()
	rows := gjson.GetBytes(body, "results.0.rows")
	elems := rows.Array()
	if rows.IsArray() && len(elems) == 0 {
		return sql.ErrNoRows
	}
	err = json.Unmarshal([]byte(elems[0].Raw), ptr)
	if err != nil {
		return errors.WithMessage(err, "json unmarshal")
	}

	return nil
}

func (d DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result := &Result{}
	resp := Response{
		Results: []*Result{result},
	}
	err := d.cli.Post("/db/execute").
		QueryParams(map[string]any{
			"timings": true,
		}).JsonRequestBody(Request(query, args...)).
		JsonResponseBody(&resp).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "call rqlite")
	}
	if result.Error != "" {
		return nil, errors.Errorf("sqlite: %s", result.Error)
	}
	return result, nil
}

func (d DB) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return nil, errors.WithMessage(err, "map to unnamed query")
	}
	return d.Exec(ctx, query, args...)
}
