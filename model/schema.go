package model

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/database"
	"isp-config-service/entity"
)

type SchemaRepository interface {
	Snapshot() ([]entity.ConfigSchema, error)
	Upsert(s entity.ConfigSchema) (*entity.ConfigSchema, error)
	Delete(identities []string) (int, error)
}

type schemaRepPg struct {
	rxClient *database.RxDbClient
}

func (r *schemaRepPg) Snapshot() ([]entity.ConfigSchema, error) {
	schemas := make([]entity.ConfigSchema, 0)
	err := r.rxClient.Visit(func(db *pg.DB) error {
		return db.Model(&schemas).Select()
	})
	return schemas, err
}

func (r *schemaRepPg) Upsert(s entity.ConfigSchema) (*entity.ConfigSchema, error) {
	err := r.rxClient.Visit(func(db *pg.DB) error {
		_, err := db.Model(&s).
			OnConflict("(id) DO UPDATE").
			Returning("*").
			Insert()
		return err
	})
	return &s, err
}

func (r *schemaRepPg) Delete(identities []string) (int, error) {
	var err error
	var res pg.Result
	err = r.rxClient.Visit(func(db *pg.DB) error {
		res, err = db.Model(&entity.ConfigSchema{}).
			Where("id IN (?)", pg.In(identities)).Delete()
		return err
	})
	return res.RowsAffected(), err
}
