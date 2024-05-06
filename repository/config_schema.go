package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
)

type ConfigSchema struct {
	db db.DB
}

func NewConfigSchema(db db.DB) ConfigSchema {
	return ConfigSchema{
		db: db,
	}
}

func (r ConfigSchema) Upsert(ctx context.Context, schema entity.ConfigSchema) error {
	query, args, err := squirrel.Insert(Table("config_schema")).
		Columns("id", "module_id", "data", "module_version").
		Values(schema.Id, schema.ModuleId, schema.Data, schema.ModuleVersion).
		Suffix(`on conflict (module_id) do update 
	set data = excluded.data, module_version = excluded.module_version, updated_at = unixepoch()`).
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
