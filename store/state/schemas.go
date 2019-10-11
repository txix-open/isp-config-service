package state

import (
	"github.com/satori/go.uuid"
	"isp-config-service/entity"
)

type SchemaStore struct {
	schemas []entity.ConfigSchema
}

func (ss *SchemaStore) Upsert(schema entity.ConfigSchema) entity.ConfigSchema {
	for key, value := range ss.schemas {
		if value.ModuleId == schema.ModuleId {
			schema.Id = value.Id
			ss.schemas[key] = schema
			return schema
		}
	}
	schema.Id = uuid.NewV1().String()
	ss.schemas = append(ss.schemas, schema)
	return schema
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
