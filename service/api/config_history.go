package api

import (
	"context"

	"isp-config-service/domain"
)

type ConfigHistory struct {
}

func (c ConfigHistory) GetAllVersion(ctx context.Context, configId string) ([]domain.ConfigVersion, error) {

}

func (c ConfigHistory) DeleteConfigVersion(ctx context.Context, id string) error {

}
