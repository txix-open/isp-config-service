package api

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type ConfigRepo interface {
	GetActive(ctx context.Context, moduleId string) (*entity.Config, error)
	GetByModuleId(ctx context.Context, moduleId string) ([]entity.Config, error)
	GetById(ctx context.Context, id string) (*entity.Config, error)
	DeleteNonActiveById(ctx context.Context, id string) (bool, error)
}

type Config struct {
	configRepo ConfigRepo
	moduleRepo ModuleRepo
	schemaRepo SchemaRepo
}

func NewConfig(
	configRepo ConfigRepo,
	moduleRepo ModuleRepo,
	schemaRepo SchemaRepo,
) Config {
	return Config{
		configRepo: configRepo,
		moduleRepo: moduleRepo,
		schemaRepo: schemaRepo,
	}
}

func (c Config) GetActiveConfigByModuleName(ctx context.Context, moduleName string) (*domain.Config, error) {
	modules, err := c.moduleRepo.GetByNames(ctx, []string{moduleName})
	if err != nil {
		return nil, errors.WithMessage(err, "get config by module name")
	}
	if len(modules) == 0 {
		return nil, entity.ErrModuleNotFound
	}

	moduleId := modules[0].Id
	config, err := c.configRepo.GetActive(ctx, moduleId)
	if err != nil {
		return nil, errors.WithMessage(err, "get active config")
	}
	if config == nil {
		return nil, entity.ErrConfigNotFound
	}

	result := configToDto(*config, nil)
	return &result, nil
}

func (c Config) GetConfigsByModuleId(ctx context.Context, moduleId string) ([]domain.Config, error) {
	configs, err := c.configRepo.GetByModuleId(ctx, moduleId)
	if err != nil {
		return nil, errors.WithMessage(err, "get configs by module id")
	}
	if len(configs) == 0 {
		return make([]domain.Config, 0), nil
	}

	schema, err := c.schemaRepo.GetByModuleId(ctx, moduleId)
	if err != nil {
		return nil, errors.WithMessage(err, "get config schema by module id")
	}
	var jsonSchema []byte
	if schema != nil {
		jsonSchema = schema.Data
	}

	result := make([]domain.Config, 0)
	for _, config := range configs {
		result = append(result, configToDto(config, jsonSchema))
	}

	return result, nil
}

func (c Config) CreateUpdateConfig(ctx context.Context, moduleId string) (*domain.Config, error) {

}

func (c Config) GetConfigById(ctx context.Context, configId string) (*domain.Config, error) {
	config, err := c.configRepo.GetById(ctx, configId)
	if err != nil {
		return nil, errors.WithMessage(err, "get config by id")
	}
	if config == nil {
		return nil, entity.ErrConfigNotFound
	}

	result := configToDto(*config, nil)
	return &result, nil
}

func (c Config) MarkConfigAsActive(ctx context.Context, configId string) error {

}

func (c Config) DeleteConfig(ctx context.Context, configId string) error {
	deleted, err := c.configRepo.DeleteNonActiveById(ctx, configId)
	if err != nil {
		return errors.WithMessage(err, "delete config")
	}
	if !deleted {
		return entity.ErrConfigNotFoundOrActive
	}
	return nil
}

func configToDto(config entity.Config, schema []byte) domain.Config {
	valid := true
	if schema != nil {
		errors, _ := validateConfig(config.Data, schema)
		valid = len(errors) == 0
	}
	return domain.Config{
		Id:        config.Id,
		Name:      config.Name,
		ModuleId:  config.ModuleId,
		Valid:     valid,
		Data:      config.Data,
		Version:   config.Version,
		Active:    bool(config.Active),
		CreatedAt: time.Time(config.CreatedAt),
		UpdatedAt: time.Time(config.UpdatedAt),
	}
}

func validateConfig(config []byte, schema []byte) (map[string]string, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schema)
	configLoader := gojsonschema.NewBytesLoader(config)
	result, err := gojsonschema.Validate(schemaLoader, configLoader)
	if err != nil {
		return nil, errors.WithMessage(err, "validate config")
	}
	details := map[string]string{}
	for _, resultError := range result.Errors() {
		details[resultError.Field()] = resultError.Description()
	}
	return details, nil
}
