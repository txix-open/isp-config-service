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

type Module struct {
	db db.DB
}

func NewModule(db db.DB) Module {
	return Module{
		db: db,
	}
}

func (r Module) Upsert(ctx context.Context, module entity.Module) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Module.Upsert")

	query, args, err := squirrel.Insert(Table("module")).
		Columns("id", "name", "last_connected_at").
		Values(module.Id, module.Name, squirrel.Expr("unixepoch()")).
		Suffix(`on conflict (name) do update set last_connected_at = unixepoch()`).
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "select: %s", query)
	}
	return nil
}

func (r Module) SetDisconnectedAtNow(
	ctx context.Context,
	moduleId string,
) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Module.SetDisconnectedAtNow")

	query, args, err := squirrel.Update(Table("module")).
		Set("last_disconnected_at", squirrel.Expr("unixepoch()")).
		Where(squirrel.Eq{"id": moduleId}).
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

func (r Module) GetByNames(ctx context.Context, names []string) ([]entity.Module, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Module.GetByNames")

	query, args, err := squirrel.Select("*").
		From(Table("module")).
		Where(squirrel.Eq{
			"name": names,
		}).OrderBy("created_at desc").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	result := make([]entity.Module, 0)
	err = r.db.Select(ctx, &result, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}

	return result, nil
}

func (r Module) GetById(ctx context.Context, id string) (*entity.Module, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Module.GetById")

	query := fmt.Sprintf("select * from %s where id = ?", Table("module"))
	return selectRow[entity.Module](ctx, r.db, query, id)
}

func (r Module) All(ctx context.Context) ([]entity.Module, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Module.All")

	result := make([]entity.Module, 0)
	query := fmt.Sprintf("select * from %s order by name", Table("module"))
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r Module) Delete(ctx context.Context, id string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Module.Delete")

	query, args, err := squirrel.Delete(Table("module")).
		Where(squirrel.Eq{"id": id}).
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

func Table(table string) string {
	return fmt.Sprintf("isp_config_service__%s", table)
}
