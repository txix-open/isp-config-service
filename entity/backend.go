package entity

import (
	"github.com/txix-open/isp-kit/cluster"
	"isp-config-service/entity/xtypes"
)

// nolint: tagliatelle
type Backend struct {
	WsConnectionId       string                                     `json:"ws_connection_id"`
	ModuleId             string                                     `json:"module_id"`
	Address              string                                     `json:"address"`
	Version              string                                     `json:"version"`
	LibVersion           string                                     `json:"lib_version"`
	ModuleName           string                                     `json:"module_name"`
	ConfigServiceNodeId  string                                     `json:"config_service_node_id"`
	Endpoints            xtypes.Json[[]cluster.EndpointDescriptor]  `json:"endpoints"`
	RequiredModules      xtypes.Json[[]cluster.ModuleDependency]    `json:"required_modules"`
	CreatedAt            xtypes.Time                                `json:"created_at"`
	MetricsAutodiscovery *xtypes.Json[cluster.MetricsAutodiscovery] `json:"metrics_autodiscovery"`
}

// nolint: tagliatelle
type MetricsAdWrapper struct {
	MetricsAutodiscovery xtypes.Json[cluster.MetricsAutodiscovery] `json:"metrics_autodiscovery"`
}
