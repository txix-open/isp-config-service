package api

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/txix-open/isp-kit/log"

	"isp-config-service/domain"
	"isp-config-service/entity"

	"github.com/pkg/errors"
)

type ConfigHistoryRepo interface {
	Delete(ctx context.Context, id string) error
	GetByConfigId(ctx context.Context, configId string) ([]entity.ConfigHistory, error)
	Insert(ctx context.Context, history entity.ConfigHistory) error
	DeleteOld(ctx context.Context, configId string, keepVersions int) (int, error)
}

type ConfigHistory struct {
	configRepo   ConfigRepo
	repo         ConfigHistoryRepo
	keepVersions int
	logger       log.Logger
}

func NewConfigHistory(
	configRepo ConfigRepo,
	repo ConfigHistoryRepo,
	keepVersions int,
	logger log.Logger,
) ConfigHistory {
	return ConfigHistory{
		repo:         repo,
		configRepo:   configRepo,
		keepVersions: keepVersions,
		logger:       logger,
	}
}

func (s ConfigHistory) GetAllVersions(ctx context.Context, configId string) ([]domain.ConfigVersion, error) {
	versions, err := s.repo.GetByConfigId(ctx, configId)
	if err != nil {
		return nil, errors.WithMessage(err, "get config versions")
	}
	actualCfg, err := s.configRepo.GetById(ctx, configId)
	if err != nil {
		return nil, errors.WithMessage(err, "get actual config by id")
	}
	if actualCfg == nil {
		return nil, entity.ErrConfigNotFound
	}

	result := make([]domain.ConfigVersion, 0, len(versions)+1)
	result = append(result, domain.ConfigVersion{
		Id:            uuid.NewString(),
		ConfigId:      actualCfg.Id,
		ConfigVersion: actualCfg.Version,
		Data:          actualCfg.Data,
		CreatedAt:     time.Time(actualCfg.UpdatedAt),
	})

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

func (s ConfigHistory) Purge(ctx context.Context, configId string, keepVersions int) (int, error) {
	deletedCount, err := s.repo.DeleteOld(ctx, configId, keepVersions)
	if err != nil {
		return 0, errors.WithMessage(err, "delete old config versions")
	}

	s.logger.Debug(ctx, fmt.Sprintf("delete '%d' old config versions", deletedCount))
	return deletedCount, nil
}

func (s ConfigHistory) OnUpdateConfig(ctx context.Context, oldConfig entity.Config) error {
	history := entity.ConfigHistory{
		Id:        uuid.NewString(),
		ConfigId:  oldConfig.Id,
		Data:      oldConfig.Data,
		Version:   oldConfig.Version,
		AdminId:   oldConfig.AdminId,
		CreatedAt: oldConfig.UpdatedAt,
	}
	err := s.repo.Insert(ctx, history)
	if err != nil {
		return errors.WithMessage(err, "save config history")
	}

	if s.keepVersions < 0 {
		return nil
	}

	go func() {
		deletedCount, err := s.repo.DeleteOld(context.Background(), oldConfig.Id, s.keepVersions)
		if err != nil {
			s.logger.Error(ctx, errors.WithMessage(err, "delete old config versions"))
		} else {
			s.logger.Debug(ctx, fmt.Sprintf("delete '%d' old config versions", deletedCount))
		}
	}()

	return nil
}
