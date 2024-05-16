package api

import (
	"context"

	"github.com/txix-open/isp-kit/grpc/apierrors"
	"isp-config-service/domain"
)

type ModuleService interface {
	Status(ctx context.Context) ([]domain.ModuleInfo, error)
	Delete(ctx context.Context, id string) error
	Connections(ctx context.Context) ([]domain.Connection, error)
}

type Module struct {
	service ModuleService
}

func NewModule(service ModuleService) Module {
	return Module{
		service: service,
	}
}

// Status
// @Summary Метод полчения полной информации о состоянии модулей
// @Description Возвращает полное состояние всех модулей в кластере (схема конфигурации, подключенные экземпляры)
// @Tags Модули
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.ModuleInfo
// @Failure 500 {object} apierrors.Error
// @Router /module/get_modules_info [POST]
func (c Module) Status(ctx context.Context) ([]domain.ModuleInfo, error) {
	modulesInfo, err := c.service.Status(ctx)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return modulesInfo, nil
}

// DeleteModule
// @Summary Метод удаления объектов модулей по идентификаторам
// @Description Возвращает количество удаленных модулей
// @Tags Модули
// @Accept  json
// @Produce  json
// @Param body body []string true "массив идентификаторов модулей"
// @Success 200 {object} domain.DeleteResponse
// @Failure 500 {object} apierrors.Error
// @Router /module/delete_module [POST]
func (c Module) DeleteModule(ctx context.Context, identities []string) (*domain.DeleteResponse, error) {
	id, err := getSingleId(identities)
	if err != nil {
		return nil, err
	}
	err = c.service.Delete(ctx, id)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}

	return &domain.DeleteResponse{
		Deleted: len(identities),
	}, nil
}

// Connections
// @Summary Метод получения маршрутов
// @Description Возвращает все доступные роуты
// @Tags Модули
// @Accept json
// @Produce json
// @Success 200 {array} domain.Connection
// @Failure 500 {object} apierrors.Error
// @Router /routing/get_routes [POST]
func (c Module) Connections(ctx context.Context) ([]domain.Connection, error) {
	connections, err := c.service.Connections(ctx)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return connections, nil
}

func getSingleId(identities []string) (string, error) {
	if len(identities) == 0 {
		return "", apierrors.NewBusinessError(domain.ErrorCodeBadRequest, "at least one id is required", nil)
	}
	if len(identities) > 1 {
		return "", apierrors.NewBusinessError(domain.ErrorCodeBadRequest, "accept only single identity", nil)
	}
	return identities[0], nil
}
