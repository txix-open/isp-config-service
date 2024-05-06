package entity

import (
	"github.com/txix-open/isp-kit/cluster"
)

type Backend struct {
	ModuleId        string                                  `json:"module_id"`
	Address         string                                  `json:"address"`
	Version         string                                  `json:"version"`
	LibVersion      string                                  `json:"lib_version"`
	Endpoints       JsonValue[[]cluster.EndpointDescriptor] `json:"endpoints"`
	RequiredModules JsonValue[cluster.ModuleRequirements]   `json:"required_modules"`
	//CreatedAt       time.Time                               `json:"created_at"`
	//UpdatedAt       time.Time                               `json:"updated_at"`
}
