package domain

import (
	"time"

	"isp-config-service/entity"

	"github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/structure"
)

type DeleteResponse struct {
	Deleted int
}

type Connection struct {
	LibVersion    string
	Version       string
	Address       structure.AddressConfiguration
	Endpoints     []structure.EndpointConfig `json:",omitempty"`
	EstablishedAt time.Time
}

type ModuleInfo struct {
	Id                 int32
	Name               string
	Active             bool
	CreatedAt          time.Time
	LastConnectedAt    time.Time
	LastDisconnectedAt time.Time
	Configs            []entity.Config `json:",omitempty"`
	ConfigSchema       *schema.Schema  `json:",omitempty"`
	Status             []Connection    `json:",omitempty"`
}
