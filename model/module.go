package model

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/database"
	"isp-config-service/entity"
)

type ModulesRepository interface {
	Snapshot() ([]entity.Module, error)
	Upsert(module entity.Module) (entity.Module, error)
	Delete(identities []int32) (int, error)
}

type modulesRepPg struct {
	rxClient *database.RxDbClient
}

func (r *modulesRepPg) Snapshot() ([]entity.Module, error) {
	modules := make([]entity.Module, 0)
	err := r.rxClient.Visit(func(db *pg.DB) error {
		return db.Model(&modules).Select()
	})
	return modules, err
}

func (r *modulesRepPg) Upsert(module entity.Module) (entity.Module, error) {
	err := r.rxClient.Visit(func(db *pg.DB) error {
		_, err := db.Model(&module).
			OnConflict("(id) DO UPDATE").
			Returning("*").
			Insert()
		return err
	})
	return module, err
}

func (r *modulesRepPg) Delete(identities []int32) (int, error) {
	var err error
	var res pg.Result
	err = r.rxClient.Visit(func(db *pg.DB) error {
		res, err = db.Model(&entity.Module{}).
			Where("id IN (?)", pg.In(identities)).Delete()
		return err
	})
	return res.RowsAffected(), err
}
