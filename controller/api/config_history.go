package api

import (
	"context"

	"isp-config-service/domain"
	"isp-config-service/entity"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
)

type ConfigHistoryService interface {
	GetAllVersions(ctx context.Context, configId string) ([]domain.ConfigVersion, error)
	Delete(ctx context.Context, id string) error
	Purge(ctx context.Context, configId string, keepVersions int) (int, error)
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
// @Param body body domain.IdRequest true "id конфигурации"
// @Success 200 {array} domain.ConfigVersion
// @Failure 400 {object} apierrors.Error "если не указан массив идентификаторов"
// @Failure 500 {object} apierrors.Error
// @Router /config/get_all_version [POST]
func (c ConfigHistory) GetAllVersion(ctx context.Context, req domain.IdRequest) ([]domain.ConfigVersion, error) {
	versions, err := c.service.GetAllVersions(ctx, req.Id)
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
		return versions, nil
	}
}

// DeleteConfigVersion
// @Summary Метод удаления версии конфигурации
// @Description Возвращает количество удаленных версий
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.IdRequest true "id версии конфигурации"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} apierrors.Error "не указан массив идентификаторов"
// @Failure 500 {object} apierrors.Error
// @Router /config/delete_version [POST]
func (c ConfigHistory) DeleteConfigVersion(ctx context.Context, req domain.IdRequest) (*domain.DeleteResponse, error) {
	err := c.service.Delete(ctx, req.Id)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return &domain.DeleteResponse{
		Deleted: 1,
	}, nil
}

// PurgeConfigVersions
// @Summary Метод удаления всех версий конфигурации, за исключением n последних
// @Description Возвращает количество удаленных версий
// @Tags Конфигурация
// @Accept json
// @Produce json
// @Param body body domain.PurgeConfigVersionsRequest true "id конфигурации, количество оставленных версий конфигурации"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} apierrors.Error "не указан массив идентификаторов"
// @Failure 500 {object} apierrors.Error
// @Router /config/purge_versions [POST]
func (c ConfigHistory) PurgeConfigVersions(ctx context.Context, req domain.PurgeConfigVersionsRequest) (*domain.DeleteResponse, error) {
	deleted, err := c.service.Purge(ctx, req.ConfigId, req.KeepVersions)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(err)
	}
	return &domain.DeleteResponse{
		Deleted: deleted,
	}, nil
}
