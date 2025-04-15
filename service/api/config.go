package api

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/txix-open/go-cmp/cmp"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/xeipuuv/gojsonschema"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
)

type ConfigRepo interface {
	GetActive(ctx context.Context, moduleId string) (*entity.Config, error)
	GetByModuleId(ctx context.Context, moduleId string) ([]entity.Config, error)
	GetById(ctx context.Context, id string) (*entity.Config, error)
	DeleteNonActiveById(ctx context.Context, id string) (bool, error)
	Insert(ctx context.Context, cfg entity.Config) error
	UpdateByVersion(ctx context.Context, cfg entity.Config) (bool, error)
	SetActive(ctx context.Context, configId string, active xtypes.Bool) error
	UpdateConfigName(ctx context.Context, req domain.UpdateConfigNameRequest) (bool, error)
}

type EventRepo interface {
	Insert(ctx context.Context, event entity.Event) error
}

type ConfigHistoryService interface {
	OnUpdateConfig(ctx context.Context, oldConfig entity.Config) error
}

type VariableService interface {
	ExtractVariables(ctx context.Context, configId string, input []byte) (*domain.VariableExtractionResult, error)
	SaveVariableLinks(ctx context.Context, configId string, links []entity.ConfigHasVariable) error
}

type Config struct {
	configRepo           ConfigRepo
	moduleRepo           ModuleRepo
	schemaRepo           SchemaRepo
	eventRepo            EventRepo
	configHistoryService ConfigHistoryService
	variableService      VariableService
	logger               log.Logger
}

func NewConfig(
	configRepo ConfigRepo,
	moduleRepo ModuleRepo,
	schemaRepo SchemaRepo,
	eventRepo EventRepo,
	configHistoryService ConfigHistoryService,
	variableService VariableService,
	logger log.Logger,
) Config {
	return Config{
		configRepo:           configRepo,
		moduleRepo:           moduleRepo,
		schemaRepo:           schemaRepo,
		eventRepo:            eventRepo,
		configHistoryService: configHistoryService,
		variableService:      variableService,
		logger:               logger,
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

func (c Config) CreateUpdateConfig(
	ctx context.Context,
	adminId int,
	req domain.CreateUpdateConfigRequest,
) (*domain.Config, error) {
	if !req.Unsafe {
		err := c.validateConfigUpdate(ctx, req.ModuleId, req.Data)
		if err != nil {
			return nil, errors.WithMessage(err, "validate config")
		}
	}

	if req.Id == "" {
		return c.insertNewConfig(ctx, adminId, req)
	}

	variableExtractionResult, err := c.extractAndCheckVariables(ctx, req.Id, req.Data)
	if err != nil {
		return nil, errors.WithMessage(err, "extract variables")
	}

	oldConfig, err := c.configRepo.GetById(ctx, req.Id)
	if err != nil {
		return nil, errors.WithMessage(err, "get old config")
	}
	if oldConfig == nil {
		return nil, entity.ErrConfigNotFound
	}

	isMatch, err := isEqualConfigs(oldConfig, req)
	if err != nil {
		return nil, errors.WithMessage(err, "compare data configs")
	}
	if isMatch {
		c.logger.Info(ctx, "no changes in config; skip update")
		result := configToDto(*oldConfig, nil)
		return &result, nil
	}

	config := entity.Config{
		Id:      req.Id,
		Name:    req.Name,
		Data:    req.Data,
		Version: req.Version,
		AdminId: adminId,
	}
	updated, err := c.configRepo.UpdateByVersion(ctx, config)
	if err != nil {
		return nil, errors.WithMessage(err, "update config")
	}
	if !updated {
		return nil, entity.ErrConfigConflictUpdate
	}

	err = c.configHistoryService.OnUpdateConfig(ctx, *oldConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "change config history")
	}

	err = c.variableService.SaveVariableLinks(ctx, config.Id, variableExtractionResult.VariableLinks)
	if err != nil {
		return nil, errors.WithMessage(err, "save variable links")
	}

	if oldConfig.Active {
		err = c.emitChangeActiveConfigEvent(ctx, oldConfig.ModuleId)
		if err != nil {
			return nil, errors.WithMessage(err, "emit change active config event")
		}
	}

	newConfig := entity.Config{
		Id:       req.Id,
		Name:     req.Name,
		ModuleId: req.ModuleId,
		Data:     req.Data,
		Version:  req.Version + 1,
		Active:   oldConfig.Active,
	}
	result := configToDto(newConfig, nil)
	return &result, nil
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
	cfg, err := c.configRepo.GetById(ctx, configId)
	if err != nil {
		return errors.WithMessage(err, "get config by id")
	}
	if cfg == nil {
		return entity.ErrConfigNotFound
	}

	currentActiveConfig, err := c.configRepo.GetActive(ctx, cfg.ModuleId)
	if err != nil {
		return errors.WithMessage(err, "get active config")
	}
	if currentActiveConfig != nil {
		err := c.configRepo.SetActive(ctx, currentActiveConfig.Id, false)
		if err != nil {
			return errors.WithMessage(err, "deactivate config")
		}
	}

	err = c.configRepo.SetActive(ctx, configId, true)
	if err != nil {
		return errors.WithMessage(err, "set active config")
	}

	err = c.emitChangeActiveConfigEvent(ctx, cfg.ModuleId)
	if err != nil {
		return errors.WithMessage(err, "emit change active config event")
	}

	return nil
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

func (c Config) UpdateConfigName(ctx context.Context, req domain.UpdateConfigNameRequest) error {
	exist, err := c.configRepo.UpdateConfigName(ctx, req)
	if err != nil {
		return errors.WithMessage(err, "update config name")
	}
	if !exist {
		return entity.ErrConfigNotFound
	}

	return nil
}

func (c Config) SyncConfig(ctx context.Context, moduleName string) error {
	modules, err := c.moduleRepo.GetByNames(ctx, []string{moduleName})
	if err != nil {
		return errors.WithMessage(err, "get config by module name")
	}
	if len(modules) == 0 {
		return entity.ErrModuleNotFound
	}

	moduleId := modules[0].Id
	err = c.emitChangeActiveConfigEvent(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "emit change active config event")
	}

	return nil
}

func (c Config) insertNewConfig(ctx context.Context, adminId int, req domain.CreateUpdateConfigRequest) (*domain.Config, error) {
	config := entity.Config{
		Id:       uuid.NewString(),
		Name:     req.Name,
		ModuleId: req.ModuleId,
		Data:     req.Data,
		Version:  1,
		Active:   false,
		AdminId:  adminId,
	}
	err := c.configRepo.Insert(ctx, config)
	if err != nil {
		return nil, errors.WithMessage(err, "insert new config")
	}

	extractResult, err := c.extractAndCheckVariables(ctx, config.Id, config.Data)
	if err != nil {
		return nil, err
	}
	err = c.variableService.SaveVariableLinks(ctx, config.Id, extractResult.VariableLinks)
	if err != nil {
		return nil, errors.WithMessage(err, "save variable links")
	}

	result := configToDto(config, nil)
	return &result, nil
}

func (c Config) extractAndCheckVariables(ctx context.Context, configId string, input []byte) (*domain.VariableExtractionResult, error) {
	variableExtractionResult, err := c.variableService.ExtractVariables(ctx, configId, input)
	if err != nil {
		return nil, errors.WithMessage(err, "extract variables")
	}
	if len(variableExtractionResult.AbsentVariables) > 0 {
		details := make(map[string]string)
		for _, variable := range variableExtractionResult.AbsentVariables {
			details[variable] = "variable is not defined"
		}
		return nil, domain.NewConfigValidationError(details)
	}
	return variableExtractionResult, nil
}

func (c Config) validateConfigUpdate(ctx context.Context, moduleId string, configData []byte) error {
	schema, err := c.schemaRepo.GetByModuleId(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "get schema by module id")
	}
	if schema == nil {
		return entity.ErrSchemaNotFound
	}

	details, err := validateConfig(configData, schema.Data)
	if err != nil {
		return errors.WithMessage(err, "validate config")
	}
	if len(details) == 0 {
		return nil
	}

	return domain.NewConfigValidationError(details)
}

func (c Config) emitChangeActiveConfigEvent(ctx context.Context, moduleId string) error {
	eventPayload := entity.EventPayload{
		ConfigUpdated: &entity.PayloadConfigUpdated{
			ModuleId: moduleId,
		},
	}
	err := c.eventRepo.Insert(ctx, entity.NewEvent(eventPayload))
	if err != nil {
		return errors.WithMessage(err, "insert new event")
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

func isEqualConfigs(oldConfig *entity.Config, reqConfig domain.CreateUpdateConfigRequest) (bool, error) {
	if oldConfig.Name != reqConfig.Name {
		return false, nil
	}

	oldDataMap := make(map[string]any)
	err := json.Unmarshal(oldConfig.Data, &oldDataMap)
	if err != nil {
		return false, errors.WithMessage(err, "unmarshal actual config data")
	}

	reqDataMap := make(map[string]any)
	err = json.Unmarshal(reqConfig.Data, &reqDataMap)
	if err != nil {
		return false, errors.WithMessage(err, "unmarshal request config data")
	}

	return cmp.Equal(oldDataMap, reqDataMap), nil
}
