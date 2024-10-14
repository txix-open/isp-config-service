package domain

import (
	"time"

	"github.com/txix-open/isp-kit/json"
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
	ConfigSchema       json.RawMessage `swaggertype:"object"`
	Status             []Connection
	CreatedAt          time.Time
}

type Connection struct {
	ModuleName    string
	LibVersion    string
	Version       string
	Address       Address
	Endpoints     []EndpointDescriptor
	EstablishedAt time.Time
}

type Address struct {
	Ip   string
	Port string
}

type EndpointDescriptor struct {
	Path  string
	Inner bool
}

type Config struct {
	Id        string
	Name      string
	ModuleId  string
	Valid     bool
	Data      json.RawMessage `swaggertype:"object"`
	Version   int
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ConfigVersion struct {
	Id            string
	ConfigId      string
	ConfigVersion int
	Data          json.RawMessage `swaggertype:"object"`
	CreatedAt     time.Time
}

type ConfigSchema struct {
	Id        string
	Version   string
	ModuleId  string
	Schema    json.RawMessage `swaggertype:"object"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
