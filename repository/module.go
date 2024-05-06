package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
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
		return "", errors.WithMessagef(err, "select : %s", query)
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

func Table(table string) string {
	return fmt.Sprintf("isp_config_service__%s", table)
}
