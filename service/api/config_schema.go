package api

import (
	"context"
	"time"

	"isp-config-service/domain"
	"isp-config-service/entity"

	"github.com/pkg/errors"
)

type ConfigSchema struct {
	schemaRepo SchemaRepo
	moduleRepo ModuleRepo
}

func NewConfigSchema(schemaRepo SchemaRepo, moduleRepo ModuleRepo) ConfigSchema {
	return ConfigSchema{
		schemaRepo: schemaRepo,
		moduleRepo: moduleRepo,
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

func (s ConfigSchema) UpdateSchemaByModuleName(ctx context.Context, request domain.UpdateSchemaRequest) error {
	modules, err := s.moduleRepo.GetByNames(ctx, []string{request.ModuleName})
	if err != nil {
		return errors.WithMessage(err, "get modules by name")
	}
	if len(modules) == 0 {
		return entity.ErrModuleNotFound
	}

	err = s.schemaRepo.UpdateByModuleId(ctx, modules[0].Id, request.Schema)
	if err != nil {
		return errors.WithMessage(err, "update schema by module id")
	}

	return nil
}
