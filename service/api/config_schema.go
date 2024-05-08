package api

import (
	"context"

	"isp-config-service/domain"
)

type ConfigSchema struct {
}

func (c ConfigSchema) SchemaByModuleId(ctx context.Context, moduleId string) (*domain.ConfigSchema, error) {

}
