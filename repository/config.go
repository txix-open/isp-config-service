package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
)

type Config struct {
	db db.DB
}

func NewConfig(db db.DB) Config {
	return Config{
		db: db,
	}
}

func (r Config) Insert(ctx context.Context, cfg entity.Config) error {
	query, args, err := squirrel.Insert(Table("config")).
		Columns("id", "name", "module_id", "data", "version", "active").
		Values(cfg.Id, cfg.Name, cfg.ModuleId, cfg.Data, cfg.Version, cfg.Active).
		Suffix("on conflict (id) do nothing").
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

func (r Config) GetActive(ctx context.Context, moduleId string) (*entity.Config, error) {
	query, args, err := squirrel.Select("*").
		From(Table("config")).
		Where(squirrel.Eq{
			"module_id": moduleId,
			"active":    "1",
		}).OrderBy("version DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	result := entity.Config{}
	err = r.db.SelectRow(ctx, &result, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrNoActiveConfig
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "select row: %s", query)
	}

	return &result, nil
}
