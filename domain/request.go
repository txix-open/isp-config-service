package domain

import (
	"github.com/txix-open/isp-kit/json"
)

type IdRequest struct {
	Id string `validate:"required"`
}

type PurgeConfigVersionsRequest struct {
	ConfigId     string `validate:"required"`
	KeepVersions int    `validate:"required"`
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
	Id            string `validate:"required"`
	NewConfigName string `validate:"required"`
}

type SyncConfigRequest struct {
	ModuleName string `validate:"required"`
}

type PrometheusTargets struct {
	Targets []string
	Labels  map[string]string
}

type AutodiscoveryResponse []PrometheusTargets
