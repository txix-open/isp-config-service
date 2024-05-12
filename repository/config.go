package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/service/rqlite/db"
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

	return selectRow[entity.Config](ctx, r.db, query, args...)
}

func (r Config) GetByModuleId(ctx context.Context, moduleId string) ([]entity.Config, error) {
	query, args, err := squirrel.Select("*").
		From(Table("config")).
		Where(squirrel.Eq{
			"module_id": moduleId,
		}).OrderBy("created_at desc").
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	result := make([]entity.Config, 0)
	err = r.db.Select(ctx, &result, query, args...)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}

	return result, nil
}

func (r Config) GetById(ctx context.Context, id string) (*entity.Config, error) {
	query, args, err := squirrel.Select("*").
		From(Table("config")).
		Where(squirrel.Eq{
			"id": id,
		}).ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	return selectRow[entity.Config](ctx, r.db, query, args...)
}

func (r Config) DeleteNonActiveById(ctx context.Context, id string) (bool, error) {
	query, args, err := squirrel.Delete(Table("config")).
		Where(squirrel.Eq{
			"id":     id,
			"active": "0",
		}).ToSql()
	if err != nil {
		return false, errors.WithMessage(err, "build query")
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, errors.WithMessagef(err, "exec: %s", query)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, errors.WithMessage(err, "get rows affected")
	}
	return affected > 0, nil
}

func (r Config) UpdateByVersion(ctx context.Context, cfg entity.Config) (bool, error) {
	query, args, err := squirrel.Update(Table("config")).
		SetMap(map[string]interface{}{
			"name":       cfg.Name,
			"data":       cfg.Data,
			"version":    squirrel.Expr("version + 1"),
			"updated_at": squirrel.Expr("unixepoch()"),
		}).Where(squirrel.Eq{
		"id":      cfg.Id,
		"version": cfg.Version,
	}).ToSql()
	if err != nil {
		return false, errors.WithMessage(err, "build query")
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, errors.WithMessagef(err, "exec: %s", query)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, errors.WithMessage(err, "get rows affected")
	}
	return affected > 0, nil
}

func (r Config) SetActive(ctx context.Context, configId string, active xtypes.Bool) error {
	query, args, err := squirrel.Update(Table("config")).
		Set("active", active).
		Set("updated_at", squirrel.Expr("unixepoch()")).
		Where(squirrel.Eq{
			"id": configId,
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
