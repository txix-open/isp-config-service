//nolint:dupl
package model

import (
	"github.com/go-pg/pg/v9"
	"github.com/integration-system/isp-lib/v2/database"
	"isp-config-service/entity"
)

type ConfigRepository interface {
	Snapshot() ([]entity.Config, error)
	Upsert(config entity.Config) (*entity.Config, error)
	Delete(identities []string) (int, error)
}

type configRepPg struct {
	rxClient *database.RxDbClient
}

func (r *configRepPg) Snapshot() ([]entity.Config, error) {
	configs := make([]entity.Config, 0)
	err := r.rxClient.Visit(func(db *pg.DB) error {
		return db.Model(&configs).Select()
	})
	return configs, err
}

func (r *configRepPg) Upsert(config entity.Config) (*entity.Config, error) {
	err := r.rxClient.Visit(func(db *pg.DB) error {
		_, err := db.Model(&config).
			OnConflict("(id) DO UPDATE").
			Returning("*").
			Insert()
		return err
	})
	if err != nil {
		return nil, err
	}
	return &config, err
}

func (r *configRepPg) Delete(identities []string) (int, error) {
	var err error
	var res pg.Result
	err = r.rxClient.Visit(func(db *pg.DB) error {
		res, err = db.Model(&entity.Config{}).
			Where("id IN (?)", pg.In(identities)).
			Delete()
		return err
	})
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), err
}
