//nolint:dupl
package model

import (
	"github.com/go-pg/pg/v9"
	"github.com/integration-system/isp-lib/v2/database"
	"isp-config-service/entity"
)

type ModulesRepository interface {
	Snapshot() ([]entity.Module, error)
	Upsert(module entity.Module) (*entity.Module, error)
	Delete(identities []string) (int, error)
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

func (r *modulesRepPg) Upsert(module entity.Module) (*entity.Module, error) {
	err := r.rxClient.Visit(func(db *pg.DB) error {
		_, err := db.Model(&module).
			OnConflict("(id) DO UPDATE").
			Returning("*").
			Insert()
		return err
	})
	if err != nil {
		return nil, err
	}
	return &module, err
}

func (r *modulesRepPg) Delete(identities []string) (int, error) {
	var err error
	var res pg.Result
	err = r.rxClient.Visit(func(db *pg.DB) error {
		res, err = db.Model(&entity.Module{}).
			Where("id IN (?)", pg.In(identities)).Delete()
		return err
	})
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), err
}
