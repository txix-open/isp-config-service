package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
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
	query, args, err := squirrel.Insert(Table("backend")).
		Columns("module_id", "address",
			"version", "lib_version",
			"endpoints", "required_modules").
		Values(backend.ModuleId, backend.Address,
			backend.Version, backend.LibVersion,
			backend.Endpoints, backend.RequiredModules,
		).Suffix(`on conflict (module_id, address)  do update
	set version = excluded.version, lib_version = excluded.lib_version,
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
