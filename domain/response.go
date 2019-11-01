package domain

import (
	"github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/entity"
	"time"
)

type DeleteResponse struct {
	Deleted int
}

type ModuleInfo struct {
	Id                 string
	Name               string
	Active             bool
	CreatedAt          time.Time
	LastConnectedAt    time.Time
	LastDisconnectedAt time.Time
	Configs            []ConfigModuleInfo `json:",omitempty"`
	ConfigSchema       *schema.Schema     `json:",omitempty"`
	Status             []Connection       `json:",omitempty"`
}

type Connection struct {
	LibVersion    string
	Version       string
	Address       structure.AddressConfiguration
	Endpoints     []structure.EndpointConfig `json:",omitempty"`
	EstablishedAt time.Time
}

type CommonConfigLinks map[string][]string

type CompiledConfigResponse map[string]interface{}

type DeleteCommonConfigResponse struct {
	Deleted bool
	Links   CommonConfigLinks
}

type CreateUpdateConfigResponse struct {
	ErrorDetails map[string]string
	Config       *ConfigModuleInfo
}

type ConfigModuleInfo struct {
	entity.Config
	Valid bool
}
