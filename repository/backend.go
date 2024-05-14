package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/middlewares/sql_metrics"
	"isp-config-service/service/rqlite/db"
)

type Backend struct {
	db db.DB
}

func NewBackend(db db.DB) Backend {
	return Backend{
		db: db,
	}
}

func (r Backend) Upsert(ctx context.Context, backend entity.Backend) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.Upsert")

	query, args, err := squirrel.Insert(Table("backend")).
		Columns("module_id", "address",
			"version", "lib_version", "module_name",
			"endpoints", "required_modules").
		Values(backend.ModuleId, backend.Address,
			backend.Version, backend.LibVersion, backend.ModuleName,
			backend.Endpoints, backend.RequiredModules,
		).Suffix(`on conflict (module_id, address)  do update
	set version = excluded.version, lib_version = excluded.lib_version, module_name = excluded.module_name,
		endpoints = excluded.endpoints, required_modules = excluded.required_modules,
		updated_at = unixepoch()
		`).
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}

	return nil
}

func (r Backend) Delete(ctx context.Context, moduleId string, address string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.Delete")

	query, args, err := squirrel.Delete(Table("backend")).
		Where(squirrel.Eq{
			"module_id": moduleId,
			"address":   address,
		}).ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}

	return nil
}

func (r Backend) All(ctx context.Context) ([]entity.Backend, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.All")

	result := make([]entity.Backend, 0)
	query := fmt.Sprintf("SELECT * FROM %s order by created_at desc", Table("backend"))
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r Backend) GetByModuleId(ctx context.Context, moduleId string) ([]entity.Backend, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.GetByModuleId")

	query, args, err := squirrel.Select("*").
		From(Table("backend")).
		Where(squirrel.Eq{
			"module_id": moduleId,
		}).OrderBy("created_at desc").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	result := make([]entity.Backend, 0)
	err = r.db.Select(ctx, &result, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}

	return result, nil
}
