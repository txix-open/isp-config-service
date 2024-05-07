package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
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

func (r Module) Upsert(ctx context.Context, module entity.Module) (string, error) {
	query, args, err := squirrel.Insert(Table("module")).
		Columns("id", "name", "last_connected_at").
		Values(module.Id, module.Name, module.LastConnectedAt).
		Suffix(`on conflict (name) do update 
	set last_connected_at = excluded.last_connected_at`).
		Suffix("returning id").
		ToSql()
	if err != nil {
		return "", errors.WithMessage(err, "build query")
	}

	result := make(map[string]string)
	err = r.db.SelectRow(ctx, &result, query, args...)
	if err != nil {
		return "", errors.WithMessagef(err, "select: %s", query)
	}
	return result["id"], nil
}

func (r Module) SetDisconnectedAt(
	ctx context.Context,
	moduleId string,
	disconnected xtypes.Time,
) error {
	query, args, err := squirrel.Update(Table("module")).
		Set("last_disconnected_at", disconnected).
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
	result := entity.Module{}
	query := fmt.Sprintf("select * from %s where id = ?", Table("module"))
	err := r.db.SelectRow(ctx, &result, query, id)
	if err != nil {
		return nil, errors.WithMessagef(err, "select row: %s", query)
	}
	return &result, nil
}

func Table(table string) string {
	return fmt.Sprintf("isp_config_service__%s", table)
}
