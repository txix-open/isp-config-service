package db

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
)

type Adapter struct {
	cli *httpcli.Client
}

func Open(ctx context.Context, client *httpcli.Client) (*Adapter, error) {
	db := &Adapter{
		cli: client,
	}

	m := map[string]any{}
	err := db.SelectRow(ctx, &m, `SELECT 1 as test`)
	if err != nil {
		return nil, errors.WithMessage(err, "ping db")
	}

	return db, nil
}

func (d Adapter) Select(ctx context.Context, ptr any, query string, args ...any) error {
	result := &Result{
		Rows: ptr,
	}
	resp := &Response{
		Results: []*Result{result},
	}
	params := map[string]any{
		"timings":     true,
		"associative": true,
	}
	consistencyFromContext(ctx).appendParams(params)
	err := d.cli.Post("/db/query").
		QueryParams(params).
		JsonRequestBody(Requests(Request(query, args...))).
		JsonResponseBody(resp).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return errors.WithMessage(err, "call rqlite")
	}
	err = catchError(resp, result)
	if err != nil {
		return err
	}

	return nil
}

func (d Adapter) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	result := &Result{}
	resp := &Response{
		Results: []*Result{result},
	}
	params := map[string]any{
		"timings":     true,
		"associative": true,
	}
	consistencyFromContext(ctx).appendParams(params)
	httpResp, err := d.cli.Post("/db/request").
		QueryParams(params).
		JsonRequestBody(Requests(Request(query, args...))).
		JsonResponseBody(resp).
		StatusCodeToError().
		Do(ctx)
	if err != nil {
		return errors.WithMessage(err, "call rqlite")
	}
	defer httpResp.Close()

	err = catchError(resp, result)
	if err != nil {
		return err
	}

	body, _ := httpResp.UnsafeBody()
	rows := gjson.GetBytes(body, "results.0.rows")
	elems := rows.Array()
	if len(elems) == 0 {
		return sql.ErrNoRows
	}
	err = json.Unmarshal([]byte(elems[0].Raw), ptr)
	if err != nil {
		return errors.WithMessage(err, "json unmarshal")
	}

	return nil
}

func (d Adapter) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	results, err := d.exec(ctx, false, Requests(Request(query, args...)))
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.Errorf("unexpected empty results")
	}
	result, ok := results[0].(*Result)
	if !ok {
		return nil, errors.Errorf("unexpected result type: %T", results[0])
	}
	err = catchError(nil, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d Adapter) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	return d.Exec(ctx, query, arg)
}

func (d Adapter) ExecTransaction(ctx context.Context, requests [][]any) ([]sql.Result, error) {
	return d.exec(ctx, true, requests)
}

func (d Adapter) exec(ctx context.Context, inTx bool, requests [][]any) ([]sql.Result, error) {
	resp := &Response{
		Results: []*Result{},
	}
	err := d.cli.Post("/db/execute").
		QueryParams(map[string]any{
			"timings":     true,
			"transaction": inTx,
		}).JsonRequestBody(requests).
		JsonResponseBody(resp).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "call rqlite")
	}
	err = catchError(resp, nil)
	if err != nil {
		return nil, err
	}

	results := make([]sql.Result, 0, len(requests))
	for _, result := range resp.Results {
		results = append(results, result)
	}

	return results, nil
}

func catchError(resp *Response, result *Result) error {
	if resp != nil && resp.Error != "" {
		return errors.Errorf("rqlite api: %s", resp.Error)
	}
	if result != nil && result.Error != "" {
		return errors.Errorf("sqlite: %s", result.Error)
	}
	return nil
}
