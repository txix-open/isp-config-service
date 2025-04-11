package domain

import (
	"github.com/txix-open/isp-kit/json"
)

type IdRequest struct {
	Id string `validate:"required"`
}

type GetByModuleIdRequest struct {
	ModuleId string `validate:"required"`
}

type GetByModuleNameRequest struct {
	ModuleName string `validate:"required"`
}

type CreateUpdateConfigRequest struct {
	Id       string
	Name     string `validate:"required"`
	ModuleId string `validate:"required"`
	Version  int
	Data     json.RawMessage `swaggertype:"object"`
	Unsafe   bool
}

type UpdateConfigNameRequest struct {
	OldConfigName string `validate:"required"`
	NewConfigName string `validate:"required"`
}
