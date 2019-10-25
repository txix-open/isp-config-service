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
	Configs            []entity.Config `json:",omitempty"`
	ConfigSchema       *schema.Schema  `json:",omitempty"`
	Status             []Connection    `json:",omitempty"`
}

type Connection struct {
	LibVersion    string
	Version       string
	Address       structure.AddressConfiguration
	Endpoints     []structure.EndpointConfig `json:",omitempty"`
	EstablishedAt time.Time
}
