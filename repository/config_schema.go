package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/service/rqlite/db"
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

func (r ConfigSchema) All(ctx context.Context) ([]entity.ConfigSchema, error) {
	result := make([]entity.ConfigSchema, 0)
	query := fmt.Sprintf("select * from %s order by created_at", Table("config_schema"))
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}
