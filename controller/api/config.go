//nolint:lll
package api

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	"google.golang.org/grpc/metadata"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type ConfigService interface {
	GetActiveConfigByModuleName(ctx context.Context, moduleName string) (*domain.Config, error)
	GetConfigsByModuleId(ctx context.Context, moduleId string) ([]domain.Config, error)
	CreateUpdateConfig(ctx context.Context, adminId int, req domain.CreateUpdateConfigRequest) (*domain.Config, error)
	GetConfigById(ctx context.Context, configId string) (*domain.Config, error)
	MarkConfigAsActive(ctx context.Context, configId string) error
	DeleteConfig(ctx context.Context, id string) error
	UpdateConfigName(ctx context.Context, req domain.UpdateConfigNameRequest) error
}

type Config struct {
	service ConfigService
}

func NewConfig(service ConfigService) Config {
	return Config{
		service: service,
	}
}

// GetActiveConfigByModuleName
// @Summary Метод получения объекта конфигурации по названию модуля
// @Description Возвращает активную конфиграцию по названию модуля
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.GetByModuleNameRequest true "название модуля"
// @Success 200 {object} domain.Config
// @Failure 400 {object} apierrors.Error "`errorCode: 2001` - модуль не найден<br/>`errorCode: 2002` - конфиг не найден"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_active_config_by_module_name [POST]
func (c Config) GetActiveConfigByModuleName(ctx context.Context, req domain.GetByModuleNameRequest) (*domain.Config, error) {
	config, err := c.service.GetActiveConfigByModuleName(ctx, req.ModuleName)
	switch {
	case errors.Is(err, entity.ErrModuleNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeModuleNotFound,
			"module not found",
			err,
		)
	case errors.Is(err, entity.ErrConfigNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeConfigNotFound,
			"active config not found",
			err,
		)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return config, nil
	}
}

// GetConfigsByModuleId
// @Summary Метод получения списка конфигураций по ID модуля
// @Description Возвращает список конфиграции по ID модуля
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.GetByModuleIdRequest true "ID модуля"
// @Success 200 {array}  domain.Config
// @Failure 400 {object} apierrors.Error "если идентификатор не указан"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_configs_by_module_id [POST]
func (c Config) GetConfigsByModuleId(ctx context.Context, req domain.GetByModuleIdRequest) ([]domain.Config, error) {
	configs, err := c.service.GetConfigsByModuleId(ctx, req.ModuleId)
	switch {
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return configs, nil
	}
}

// CreateUpdateConfig
// @Summary Метод обновления конфигурации
// @Description Если конфиг с таким id существует, то обновляет данные, если нет, то добавляет данные в базу
// @Description В случае обновления рассылает всем подключенным модулям актуальную конфигурацию
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.CreateUpdateConfigRequest true "объект для сохранения"
// @Success 200 {object} domain.Config
// @Failure 400 {object} apierrors.Error "`errorCode: 2003` - конфиг не соотвествует текущей схеме<br/>`errorCode: 2002` - указанного id не сущесвует<br/>`errorCode: 2004` - кто-то уже обновил конфигурацию<br/>`errorCode: 2005` - схема конфигурации не найдена<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /config/create_update_config [POST]
func (c Config) CreateUpdateConfig(
	ctx context.Context,
	authData grpc.AuthData,
	req domain.CreateUpdateConfigRequest,
) (*domain.Config, error) {
	adminIdValue, _ := grpc.StringFromMd("x-admin-id", metadata.MD(authData))
	var adminId int
	if adminIdValue != "" {
		adminId, _ = strconv.Atoi(adminIdValue)
	}

	config, err := c.service.CreateUpdateConfig(ctx, adminId, req)

	var validationError domain.ConfigValidationError
	switch {
	case errors.As(err, &validationError):
		details := map[string]any{}
		for key, value := range validationError.Details {
			details[key] = value
		}
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeInvalidConfig,
			"invalid config",
			err,
		).WithDetails(details)
	case errors.Is(err, entity.ErrConfigNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeConfigNotFound,
			"config not found",
			err,
		)
	case errors.Is(err, entity.ErrConfigConflictUpdate):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeConfigVersionConflict,
			"someone has updated the config",
			err,
		)
	case errors.Is(err, entity.ErrSchemaNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeSchemaNotFound,
			"config schema not found",
			err,
		)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return config, nil
	}
}

// GetConfigById
// @Summary Метод получение конфигурации по id
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.IdRequest true "id конфигурации"
// @Success 200 {object} domain.Config
// @Failure 400 {object} apierrors.Error "`errorCode: 2002` - конфиг не найден<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_config_by_id [POST]
func (c Config) GetConfigById(ctx context.Context, req domain.IdRequest) (*domain.Config, error) {
	config, err := c.service.GetConfigById(ctx, req.Id)
	switch {
	case errors.Is(err, entity.ErrConfigNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeConfigNotFound,
			"config not found",
			err,
		)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return config, nil
	}
}

// MarkConfigAsActive
// @Summary Метод активации конфигурации для модуля
// @Description Активирует указанную конфигурацию и деактивирует остальные, возвращает активированную конфигурацию
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.IdRequest true "id конфигурации для изменения"
// @Success 200
// @Failure 400 {object} apierrors.Error "`errorCode: 2002` - конфиг не найден<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /config/mark_config_as_active [POST]
func (c Config) MarkConfigAsActive(ctx context.Context, identity domain.IdRequest) error {
	err := c.service.MarkConfigAsActive(ctx, identity.Id)
	switch {
	case errors.Is(err, entity.ErrConfigNotFound):
		return apierrors.NewBusinessError(
			domain.ErrorCodeConfigNotFound,
			"config not found",
			err,
		)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	default:
		return nil
	}
}

// DeleteConfigs
// @Summary Метод удаления объектов конфигурации по идентификаторам
// @Description Возвращает количество удаленных модулей
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body []string true "массив идентификаторов конфигураций"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} apierrors.Error "если не указан массив идентификаторов"
// @Failure 500 {object} apierrors.Error
// @Router /config/delete_config [POST]
func (c Config) DeleteConfigs(ctx context.Context, identities []string) (*domain.DeleteResponse, error) {
	configId, err := getSingleId(identities)
	if err != nil {
		return nil, err
	}

	err = c.service.DeleteConfig(ctx, configId)
	switch {
	case errors.Is(err, entity.ErrConfigNotFoundOrActive):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeConfigNotFound,
			"config not found or is active",
			err,
		)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return &domain.DeleteResponse{
			Deleted: 1,
		}, nil
	}
}

func (c Config) UpdateConfigName(ctx context.Context, req domain.UpdateConfigNameRequest) error {
	err := c.service.UpdateConfigName(ctx, req)
	switch {
	case errors.Is(err, entity.ErrConfigNotFound):
		return apierrors.NewBusinessError(
			domain.ErrorCodeConfigNotFound,
			"config not found",
			err,
		)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	default:
		return nil
	}
}
