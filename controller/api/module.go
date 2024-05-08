package api

import (
	"context"

	"github.com/txix-open/isp-kit/grpc/apierrors"
	"isp-config-service/domain"
)

type ModuleService interface {
	Status(ctx context.Context) ([]domain.ModuleInfo, error)
	Delete(ctx context.Context, idList []string) error
}

type Module struct {
	service ModuleService
}

func NewModule(service ModuleService) Module {
	return Module{
		service: service,
	}
}

// GetModulesAggregatedInfo
// @Summary Метод полчения полной информации о состоянии модулей
// @Description Возвращает полное состояние всех модулей в кластере (схема конфигурации, подключенные экземпляры)
// @Tags Модули
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.ModuleInfo
// @Failure 500 {object} apierrors.Error
// @Router /module/get_modules_info [POST]
func (c Module) GetModulesAggregatedInfo(ctx context.Context) ([]domain.ModuleInfo, error) {
	modulesInfo, err := c.service.Status(ctx)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return modulesInfo, nil
}

// DeleteModules
// @Summary Метод удаления объектов модулей по идентификаторам
// @Description Возвращает количество удаленных модулей
// @Tags Модули
// @Accept  json
// @Produce  json
// @Param body body []string true "массив идентификаторов модулей"
// @Success 200 {object} domain.DeleteResponse
// @Failure 500 {object} apierrors.Error
// @Router /module/delete_module [POST]
func (c Module) DeleteModules(ctx context.Context, identities []string) (*domain.DeleteResponse, error) {
	if len(identities) == 0 {
		return nil, apierrors.NewBusinessError(domain.ErrorCodeBadRequest, "at least one id is required", nil)
	}

	err := c.service.Delete(ctx, identities)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}

	return &domain.DeleteResponse{
		Deleted: len(identities),
	}, nil
}
