package api

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
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
// @Failure 400 {object} apierrors.Error "если конфигурация не найдена"
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
			domain.ErrorCodeModuleNotFound,
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
// @Failure 404 {object} apierrors.Error "если конфигурация не найдена"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_configs_by_module_id [POST]
func (c Config) GetConfigsByModuleId(ctx context.Context, req domain.GetByModuleIdRequest) ([]domain.Config, error) {
	c.service.GetConfigsByModuleId(, re)
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
// @Failure 404 {object} apierrors.Error "если конфигурация не найдена"
// @Failure 500 {object} apierrors.Error
// @Router /config/create_update_config [POST]
func (c Config) CreateUpdateConfig(ctx context.Context, config domain.CreateUpdateConfigRequest) (*domain.Config, error) {

}

// GetConfigById
// @Summary Метод получение актуальной конфигурации конфигурации
// @Description Возвращает актуальную версию конфигурации без дополнительного содержимого (ConfigData)
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.ConfigIdRequest true "id конфигурации"
// @Success 200 {object} domain.Config
// @Failure 400 {object} apierrors.Error "если не указан идентификатор конфигурации"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_config_by_id [POST]
func (c Config) GetConfigById(ctx context.Context, req domain.ConfigIdRequest) (domain.Config, error) {

}

// MarkConfigAsActive
// @Summary Метод активации конфигурации для модуля
// @Description Активирует указанную конфигурацию и деактивирует остальные, возвращает активированную конфигурацию
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.ConfigIdRequest true "id конфигурации для изменения"
// @Success 200 {object} domain.Config "активированная конфигурация"
// @Failure 404 {object} apierrors.Error "если конфигурация не найдена"
// @Failure 500 {object} apierrors.Error
// @Router /config/mark_config_as_active [POST]
func (c Config) MarkConfigAsActive(ctx context.Context, identity domain.ConfigIdRequest) (*domain.Config, error) {

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

}
