package entity

import (
	"time"
)

type Config struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	ModuleId  string    `json:"module_id"`
	Data      []byte    `json:"data"`
	Version   int       `json:"version"`
	Active    int       `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ConfigHistory struct {
	Id        string    `json:"id"`
	ConfigId  string    `json:"config_id"`
	Data      []byte    `json:"data"`
	Version   int       `json:"version"`
	AdminId   int       `json:"admin_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ConfigSchema struct {
	Id            string    `json:"id"`
	ModuleId      string    `json:"module_id"`
	Data          []byte    `json:"data"`
	ModuleVersion string    `json:"module_version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
