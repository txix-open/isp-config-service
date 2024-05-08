package api

import (
	"context"

	"isp-config-service/domain"
)

type Config struct {
}

func (c Config) GetActiveConfigByModuleName(ctx context.Context, moduleName string) (*domain.Config, error) {

}

func (c Config) GetConfigsByModuleId(ctx context.Context, moduleId string) ([]domain.Config, error) {

}

func (c Config) CreateUpdateConfig(ctx context.Context, moduleId string) (*domain.Config, error) {

}

func (c Config) GetConfigById(ctx context.Context, configId string) (*domain.Config, error) {

}

func (c Config) MarkConfigAsActive(ctx context.Context, configId string) error {

}

func (c Config) DeleteConfigs(ctx context.Context, idList []string) error {

}
