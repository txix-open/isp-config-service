package api

import (
	"context"

	"github.com/txix-open/isp-kit/grpc/apierrors"
	"isp-config-service/domain"
)

type ConfigHistoryService interface {
	GetAllVersions(ctx context.Context, configId string) ([]domain.ConfigVersion, error)
	Delete(ctx context.Context, id string) error
}

type ConfigHistory struct {
	service ConfigHistoryService
}

func NewConfigHistory(service ConfigHistoryService) ConfigHistory {
	return ConfigHistory{
		service: service,
	}
}

// GetAllVersion
// @Summary Метод получение старых версий конфигурации
// @Description Возвращает предыдущие версии конфигураций
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.ConfigIdRequest true "id конфигурации"
// @Success 200 {array} domain.ConfigVersion
// @Failure 400 {object} apierrors.Error "если не указан массив идентификаторов"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_all_version [POST]
func (c ConfigHistory) GetAllVersion(ctx context.Context, req domain.ConfigIdRequest) ([]domain.ConfigVersion, error) {
	versions, err := c.service.GetAllVersions(ctx, req.Id)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return versions, nil
}

// DeleteConfigVersion
// @Summary Метод удаления версии конфигурации
// @Description Возвращает количество удаленных версий
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.ConfigIdRequest true "id версии конфигурации"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} apierrors.Error "не указан массив идентификаторов"
// @Failure 500 {object} apierrors.Error
// @Router /config/delete_version [POST]
func (c ConfigHistory) DeleteConfigVersion(ctx context.Context, req domain.ConfigIdRequest) (*domain.DeleteResponse, error) {
	err := c.service.Delete(ctx, req.Id)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return &domain.DeleteResponse{
		Deleted: 1,
	}, nil
}
