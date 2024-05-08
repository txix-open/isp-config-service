package domain

import (
	"time"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/rc/schema"
)

type DeleteResponse struct {
	Deleted int
}

type ModuleInfo struct {
	Id                 string
	Name               string
	Active             bool
	LastConnectedAt    *time.Time
	LastDisconnectedAt *time.Time
	ConfigSchema       json.RawMessage
	Status             []Connection
	CreatedAt          time.Time
}

type Connection struct {
	LibVersion    string
	Version       string
	Address       cluster.AddressConfiguration
	Endpoints     []cluster.EndpointDescriptor
	EstablishedAt time.Time
}

type Config struct {
	Id        string
	Name      string
	ModuleId  string
	Valid     bool
	Data      json.RawMessage
	Version   int
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ConfigVersion struct {
	Id            string
	ConfigId      string
	ConfigVersion int32
	Data          json.RawMessage
	CreatedAt     time.Time
}

type ConfigSchema struct {
	Id        string
	Version   string
	ModuleId  string
	Schema    schema.Schema
	CreatedAt time.Time
	UpdatedAt time.Time
}
