package api

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type ConfigSchema struct {
	schemaRepo SchemaRepo
}

func NewConfigSchema(schemaRepo SchemaRepo) ConfigSchema {
	return ConfigSchema{
		schemaRepo: schemaRepo,
	}
}

func (s ConfigSchema) SchemaByModuleId(ctx context.Context, moduleId string) (*domain.ConfigSchema, error) {
	schema, err := s.schemaRepo.GetByModuleId(ctx, moduleId)
	if err != nil {
		return nil, errors.WithMessage(err, "get config schema")
	}
	if schema == nil {
		return nil, entity.ErrSchemaNotFound
	}

	result := domain.ConfigSchema{
		Id:        schema.Id,
		Version:   schema.ModuleVersion,
		ModuleId:  schema.ModuleId,
		Schema:    schema.Data,
		CreatedAt: time.Time(schema.CreatedAt),
		UpdatedAt: time.Time(schema.UpdatedAt),
	}

	return &result, nil
}
