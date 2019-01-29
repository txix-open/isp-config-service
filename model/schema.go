package model

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"isp-config-service/entity"
)

type SchemaRepository struct {
	DB orm.DB
}

func (r *SchemaRepository) InsertConfigSchema(s *entity.ConfigSchema) (*entity.ConfigSchema, error) {
	_, err := r.DB.Model(s).Returning("*").Insert()
	return s, err
}

func (r *SchemaRepository) UpdateConfigSchema(s *entity.ConfigSchema) (*entity.ConfigSchema, error) {
	_, err := r.DB.Model(s).Returning("*").WherePK().Update()
	return s, err
}

func (r *SchemaRepository) GetSchemasByModulesId(modulesId []int32) ([]entity.ConfigSchema, error) {
	var res []entity.ConfigSchema
	err := r.DB.Model(&res).Where("module_id IN (?)", pg.In(modulesId)).Select()
	if err != nil {
		return nil, err
	}
	return res, nil
}
