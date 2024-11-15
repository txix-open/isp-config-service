package main

import (
	"time"

	"github.com/txix-open/isp-kit/json"
)

type Module struct {
	Id                 string
	Name               string
	LastConnectedAt    *time.Time
	LastDisconnectedAt *time.Time
	CreatedAt          time.Time
}

type Config struct {
	Id        string
	ModuleId  string
	Name      string
	Version   int
	Active    bool
	Data      json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ConfigHistory struct {
	Id            string
	ConfigId      string
	ConfigVersion int
	Data          json.RawMessage
	CreatedAt     time.Time
}

type ConfigSchema struct {
	Id        string
	ModuleId  string
	Version   string
	Schema    json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}
