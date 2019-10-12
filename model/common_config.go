package model

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/database"
	"isp-config-service/entity"
)

type CommonConfigRepository interface {
	Snapshot() ([]entity.CommonConfig, error)
	Upsert(config *entity.CommonConfig) (*entity.CommonConfig, error)
	Delete(identities []int64) (int, error)
}

type commonConfigRepPg struct {
	rxClient *database.RxDbClient
}

func (r *commonConfigRepPg) Snapshot() ([]entity.CommonConfig, error) {
	configs := make([]entity.CommonConfig, 0)
	err := r.rxClient.Visit(func(db *pg.DB) error {
		return db.Model(&configs).Select()
	})
	return configs, err
}

func (r *commonConfigRepPg) Upsert(config *entity.CommonConfig) (*entity.CommonConfig, error) {
	err := r.rxClient.Visit(func(db *pg.DB) error {
		_, err := db.Model(config).
			OnConflict("(id) DO UPDATE").
			Returning("*").
			Insert()
		return err
	})
	return config, err
}

func (r *commonConfigRepPg) Delete(identities []int64) (int, error) {
	var err error
	var res pg.Result
	err = r.rxClient.Visit(func(db *pg.DB) error {
		res, err = db.Model(&entity.CommonConfig{}).
			Where("id IN (?)", pg.In(identities)).
			Delete()
		return err
	})
	return res.RowsAffected(), err
}
