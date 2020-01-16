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
	Schemas []entity.ConfigSchema
}

func (ss SchemaStore) GetByModuleIds(ids []string) []entity.ConfigSchema {
	idsMap := StringsToMap(ids)
	result := make([]entity.ConfigSchema, 0, len(ids))
	for _, schema := range ss.Schemas {
		if _, ok := idsMap[schema.ModuleId]; ok {
			result = append(result, schema)
		}
	}
	return result
}

func (ss *SchemaStore) DeleteByIds(ids []string) int {
	idsMap := StringsToMap(ids)
	var deleted int
	for i := 0; i < len(ss.Schemas); i++ {
		id := ss.Schemas[i].Id
		if _, ok := idsMap[id]; ok {
			// change schemas ordering
			ss.Schemas[i] = ss.Schemas[len(ss.Schemas)-1]
			ss.Schemas = ss.Schemas[:len(ss.Schemas)-1]
			deleted++
		}
	}
	return deleted
}

func (ss *SchemaStore) Upsert(schema entity.ConfigSchema) entity.ConfigSchema {
	for key, value := range ss.Schemas {
		if value.ModuleId == schema.ModuleId {
			schema.Id = value.Id
			schema.CreatedAt = value.CreatedAt
			ss.Schemas[key] = schema
			return schema
		}
	}
	ss.Schemas = append(ss.Schemas, schema)
	return schema
}

func NewSchemaStore() *SchemaStore {
	return &SchemaStore{Schemas: make([]entity.ConfigSchema, 0)}
}
