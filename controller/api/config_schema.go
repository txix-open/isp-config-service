package api

import (
	"context"

	"isp-config-service/domain"
	"isp-config-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
)

type ConfigSchemaService interface {
	SchemaByModuleId(ctx context.Context, moduleId string) (*domain.ConfigSchema, error)
	UpdateSchemaByModuleName(ctx context.Context, request domain.UpdateSchemaRequest) error
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
// @Failure 400 {object} apierrors.Error "`errorCode: 2005` - схема для модуля не найдена"
// @Failure 500 {object} apierrors.Error
// @Router /schema/get_by_module_id [POST]
func (c ConfigSchema) SchemaByModuleId(ctx context.Context, request domain.GetByModuleIdRequest) (*domain.ConfigSchema, error) {
	schema, err := c.service.SchemaByModuleId(ctx, request.ModuleId)
	switch {
	case errors.Is(err, entity.ErrSchemaNotFound):
		return nil, apierrors.NewBusinessError(
			domain.ErrorCodeSchemaNotFound,
			"config schema not found",
			err,
		)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return schema, nil
	}
}

// UpdateSchemaByModuleName
// @Summary Метод обновления схемы конфигурации модуля
// @Description Меняет схему конфигурации модуля на переданную
// @Tags Схема
// @Accept json
// @Produce json
// @Param body body domain.UpdateSchemaRequest true "название модуля и схема конфига"
// @Success 200 {object} any
// @Failure 400 {object} apierrors.Error "`errorCode: 2001` - модуль не найден"
// @Failure 500 {object} apierrors.Error
// @Router /schema/update_by_module_name [POST]
func (c ConfigSchema) UpdateSchemaByModuleName(ctx context.Context, request domain.UpdateSchemaRequest) error {
	err := c.service.UpdateSchemaByModuleName(ctx, request)
	switch {
	case errors.Is(err, entity.ErrModuleNotFound):
		return apierrors.NewBusinessError(
			domain.ErrorCodeModuleNotFound,
			"module not found",
			err,
		)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	default:
		return nil
	}
}
