package model

import (
	"github.com/go-pg/pg/v9"
	"github.com/integration-system/isp-lib/v2/database"
	"isp-config-service/entity"
)

type VersionConfigRepository interface {
	Snapshot() ([]entity.VersionConfig, error)
	Upsert(model entity.VersionConfig) (*entity.VersionConfig, error)
	Delete(identities string) (int, error)
}

type versionConfigRepPg struct {
	rxClient *database.RxDbClient
}

func (r *versionConfigRepPg) Snapshot() ([]entity.VersionConfig, error) {
	modules := make([]entity.VersionConfig, 0)
	err := r.rxClient.Visit(func(db *pg.DB) error {
		return db.Model(&modules).Select()
	})
	return modules, err
}

func (r *versionConfigRepPg) Upsert(module entity.VersionConfig) (*entity.VersionConfig, error) {
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

func (r *versionConfigRepPg) Delete(id string) (int, error) {
	var err error
	var res pg.Result
	err = r.rxClient.Visit(func(db *pg.DB) error {
		res, err = db.Model(&entity.VersionConfig{}).
			Where("id = ?", id).Delete()
		return err
	})
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), err
}
