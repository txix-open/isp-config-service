package api

import (
	"context"

	"isp-config-service/domain"
)

type ConfigSchemaService interface {
	SchemaByModuleId(ctx context.Context, moduleId string) (*domain.ConfigSchema, error)
}

type ConfigSchema struct {
	service ConfigSchemaService
}

func NewConfigSchema(service ConfigSchemaService) ConfigSchema {
	return ConfigSchema{
		service: service,
	}
}

// SchemaByModuleId
// @Summary Метод получения схемы конфигурации модуля
// @Description Возвращает текущую json схему конфигурации модуля
// @Tags Схема
// @Accept json
// @Produce json
// @Param body body domain.GetByModuleIdRequest true "идентификатор модуля"
// @Success 200 {object} domain.ConfigSchema
// @Failure 404 {object} apierrors.Error "если схема для модуля не найдена"
// @Failure 500 {object} apierrors.Error
// @Router /schema/get_by_module_id [POST]
func (c ConfigSchema) SchemaByModuleId(ctx context.Context, request domain.GetByModuleIdRequest) (*domain.ConfigSchema, error) {

}
