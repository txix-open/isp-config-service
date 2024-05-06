package entity

import (
	"github.com/pkg/errors"
	"isp-config-service/entity/xtypes"
)

var (
	ErrNoActiveConfig = errors.New("no active config")
)

type Config struct {
	Id        string      `json:"id"`
	Name      string      `json:"name"`
	ModuleId  string      `json:"module_id"`
	Data      []byte      `json:"data"`
	Version   int         `json:"version"`
	Active    xtypes.Bool `json:"active"`
	CreatedAt xtypes.Time `json:"created_at"`
	UpdatedAt xtypes.Time `json:"updated_at"`
}

type ConfigHistory struct {
	Id        string      `json:"id"`
	ConfigId  string      `json:"config_id"`
	Data      []byte      `json:"data"`
	Version   int         `json:"version"`
	AdminId   int         `json:"admin_id"`
	CreatedAt xtypes.Time `json:"created_at"`
}

type ConfigSchema struct {
	Id            string      `json:"id"`
	ModuleId      string      `json:"module_id"`
	Data          []byte      `json:"data"`
	ModuleVersion string      `json:"module_version"`
	CreatedAt     xtypes.Time `json:"created_at"`
	UpdatedAt     xtypes.Time `json:"updated_at"`
}
