package state

import "isp-config-service/entity"

type SchemaStore struct {
	schemas []entity.ConfigSchema
}

func (ss *SchemaStore) Upsert(schema entity.ConfigSchema) {
	// TODO
}

func (ss *SchemaStore) GetByModuleId(moduleId int32) (*entity.ConfigSchema, error) {
	// TODO
	return nil, nil
}

func (ss *SchemaStore) GetById(id int64) (*entity.ConfigSchema, error) {
	// TODO
	return nil, nil
}

func NewSchemaStore() *SchemaStore {
	return &SchemaStore{schemas: make([]entity.ConfigSchema, 0)}
}
