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

func (r Backend) Insert(ctx context.Context, backend entity.Backend) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.Insert")

	query, args, err := squirrel.Insert(Table("backend")).
		Columns("ws_connection_id", "module_id", "address",
			"version", "lib_version", "module_name", "config_service_node_id",
			"endpoints", "required_modules", "metrics_autodiscovery").
		Values(backend.WsConnectionId, backend.ModuleId, backend.Address,
			backend.Version, backend.LibVersion, backend.ModuleName, backend.ConfigServiceNodeId,
			backend.Endpoints, backend.RequiredModules, backend.MetricsAutodiscovery,
		).
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

func (r Backend) DeleteByWsConnectionIds(ctx context.Context, wsConnectionIds []string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.DeleteByWsConnectionIds")

	query, args, err := squirrel.Delete(Table("backend")).
		Where(squirrel.Eq{
			"ws_connection_id": wsConnectionIds,
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

func (r Backend) DeleteByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) (int, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.DeleteByConfigServiceNodeId")

	query, args, err := squirrel.Delete(Table("backend")).
		Where(squirrel.Eq{
			"config_service_node_id": configServiceNodeId,
		}).ToSql()
	if err != nil {
		return 0, errors.WithMessage(err, "build query")
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, errors.WithMessagef(err, "exec: %s", query)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WithMessage(err, "rows affected")
	}

	return int(deleted), nil
}

func (r Backend) GetByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) ([]entity.Backend, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.GetByConfigServiceNodeId")

	query, args, err := squirrel.Select("*").
		From(Table("backend")).
		Where(squirrel.Eq{
			"config_service_node_id": configServiceNodeId,
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

func (r Backend) AllMetricsAutodiscovery(ctx context.Context) ([]entity.MetricsAdWrapper, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Backend.AllMetricsAutodiscovery")

	result := make([]entity.MetricsAdWrapper, 0)
	query := fmt.Sprintf(`
		SELECT metrics_autodiscovery FROM %s
		WHERE metrics_autodiscovery IS NOT NULL
        ORDER BY module_name, created_at DESC`,
		Table("backend"),
	)
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select %s", query)
	}

	return result, nil
}
