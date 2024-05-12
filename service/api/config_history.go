package api

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type ConfigHistoryRepo interface {
	Delete(ctx context.Context, id string) error
	GetByConfigId(ctx context.Context, configId string) ([]entity.ConfigHistory, error)
	Insert(ctx context.Context, history entity.ConfigHistory) error
}

type ConfigHistory struct {
	repo ConfigHistoryRepo
}

func NewConfigHistory(repo ConfigHistoryRepo) ConfigHistory {
	return ConfigHistory{
		repo: repo,
	}
}

func (s ConfigHistory) GetAllVersions(ctx context.Context, configId string) ([]domain.ConfigVersion, error) {
	versions, err := s.repo.GetByConfigId(ctx, configId)
	if err != nil {
		return nil, errors.WithMessage(err, "get config versions")
	}

	result := make([]domain.ConfigVersion, 0, len(versions))
	for _, version := range versions {
		result = append(result, domain.ConfigVersion{
			Id:            version.Id,
			ConfigId:      version.ConfigId,
			ConfigVersion: version.Version,
			Data:          version.Data,
			CreatedAt:     time.Time(version.CreatedAt),
		})
	}

	return result, nil
}

func (s ConfigHistory) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return errors.WithMessage(err, "delete config version")
	}
	return nil
}
