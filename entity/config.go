package entity

import (
	"time"

	"github.com/integration-system/isp-lib/config/schema"
)

type ConfigData map[string]interface{}

type Config struct {
	//nolint
	tableName     string     `sql:"?db_schema.configs" json:"-"`
	Id            string     `json:"id"`
	Name          string     `json:"name" valid:"required~Required"`
	CommonConfigs []string   `json:"commonConfigs" pg:",array"`
	Description   string     `json:"description"`
	ModuleId      string     `json:"moduleId" valid:"required~Required"`
	Version       int32      `json:"version" sql:",null"`
	Active        bool       `json:"active" sql:",null"`
	CreatedAt     time.Time  `json:"createdAt" sql:",null"`
	UpdatedAt     time.Time  `json:"updatedAt" sql:",null"`
	Data          ConfigData `json:"data" sql:",notnull"`
}

type CommonConfig struct {
	//nolint
	tableName   string     `sql:"?db_schema.common_configs" json:"-"`
	Id          string     `json:"id"`
	Name        string     `json:"name" valid:"required~Required"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt" sql:",null"`
	UpdatedAt   time.Time  `json:"updatedAt" sql:",null"`
	Data        ConfigData `json:"data" sql:",notnull"`
}

type ConfigSchema struct {
	//nolint
	tableName string `sql:"?db_schema.config_schemas" json:"-"`
	Id        string `json:"id"`
	Version   string
	ModuleId  string
	Schema    schema.Schema
	CreatedAt time.Time
	UpdatedAt time.Time
}
