package state

import (
	"isp-config-service/entity"
)

type WriteableSchemaStore interface {
	ReadonlySchemaStore
	Upsert(schema entity.ConfigSchema) entity.ConfigSchema
	DeleteByIds(ids []string) int
}

type ReadonlySchemaStore interface {
	GetByModuleIds(ids []string) []entity.ConfigSchema
}

type SchemaStore struct {
	schemas []entity.ConfigSchema
}

func (ss SchemaStore) GetByModuleIds(ids []string) []entity.ConfigSchema {
	idsMap := StringsToMap(ids)
	result := make([]entity.ConfigSchema, 0, len(ids))
	for _, schema := range ss.schemas {
		if _, ok := idsMap[schema.ModuleId]; ok {
			result = append(result, schema)
		}
	}
	return result
}

func (ss *SchemaStore) DeleteByIds(ids []string) int {
	idsMap := StringsToMap(ids)
	var deleted int
	for i := 0; i < len(ss.schemas); i++ {
		id := ss.schemas[i].Id
		if _, ok := idsMap[id]; ok {
			// change schemas ordering
			ss.schemas[i] = ss.schemas[len(ss.schemas)-1]
			ss.schemas = ss.schemas[:len(ss.schemas)-1]
			deleted++
		}
	}
	return deleted
}

func (ss *SchemaStore) Upsert(schema entity.ConfigSchema) entity.ConfigSchema {
	for key, value := range ss.schemas {
		if value.ModuleId == schema.ModuleId {
			schema.Id = value.Id
			schema.CreatedAt = value.CreatedAt
			ss.schemas[key] = schema
			return schema
		}
	}
	ss.schemas = append(ss.schemas, schema)
	return schema
}

func NewSchemaStore() *SchemaStore {
	return &SchemaStore{schemas: make([]entity.ConfigSchema, 0)}
}
