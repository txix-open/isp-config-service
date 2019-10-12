package state

import (
	"isp-config-service/entity"
)

type WriteableSchemaStore interface {
	ReadonlySchemaStore
	Upsert(schema entity.ConfigSchema) entity.ConfigSchema
}

type ReadonlySchemaStore interface {
	GetByModuleId(moduleId int32) *entity.ConfigSchema
	GetById(id int64) *entity.ConfigSchema
}

type SchemaStore struct {
	schemas []entity.ConfigSchema
}

func (ss SchemaStore) GetByModuleId(moduleId int32) *entity.ConfigSchema {
	// TODO
	return nil
}

func (ss SchemaStore) GetById(id int64) *entity.ConfigSchema {
	// TODO
	return nil
}

func (ss *SchemaStore) Upsert(schema entity.ConfigSchema) entity.ConfigSchema {
	for key, value := range ss.schemas {
		if value.ModuleId == schema.ModuleId {
			schema.Id = value.Id
			ss.schemas[key] = schema
			return schema
		}
	}
	schema.Id = GenerateId()
	ss.schemas = append(ss.schemas, schema)
	return schema
}

func NewSchemaStore() *SchemaStore {
	return &SchemaStore{schemas: make([]entity.ConfigSchema, 0)}
}
