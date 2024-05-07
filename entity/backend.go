package entity

import (
	"github.com/txix-open/isp-kit/cluster"
	"isp-config-service/entity/xtypes"
)

type Backend struct {
	ModuleId        string                                    `json:"module_id"`
	Address         string                                    `json:"address"`
	Version         string                                    `json:"version"`
	LibVersion      string                                    `json:"lib_version"`
	ModuleName      string                                    `json:"module_name"`
	Endpoints       xtypes.Json[[]cluster.EndpointDescriptor] `json:"endpoints"`
	RequiredModules xtypes.Json[[]cluster.ModuleDependency]   `json:"required_modules"`
	CreatedAt       xtypes.Time                               `json:"created_at"`
	UpdatedAt       xtypes.Time                               `json:"updated_at"`
}
