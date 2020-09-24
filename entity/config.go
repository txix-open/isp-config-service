package entity

import (
	"time"

	"github.com/integration-system/isp-lib/v2/config/schema"
)

type ConfigData map[string]interface{}

type Config struct {
	//nolint
	tableName     string     `pg:"?db_schema.configs" json:"-"`
	Id            string     `json:"id"`
	Name          string     `json:"name" valid:"required~Required"`
	CommonConfigs []string   `json:"commonConfigs" pg:",array"`
	Description   string     `json:"description"`
	ModuleId      string     `json:"moduleId" valid:"required~Required"`
	Version       int32      `json:"version" pg:",null"`
	Active        bool       `json:"active" pg:",null"`
	CreatedAt     time.Time  `json:"createdAt" pg:",null"`
	UpdatedAt     time.Time  `json:"updatedAt" pg:",null"`
	Data          ConfigData `json:"data,omitempty" pg:",notnull"`
}

type CommonConfig struct {
	//nolint
	tableName   string     `pg:"?db_schema.common_configs" json:"-"`
	Id          string     `json:"id"`
	Name        string     `json:"name" valid:"required~Required"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt" pg:",null"`
	UpdatedAt   time.Time  `json:"updatedAt" pg:",null"`
	Data        ConfigData `json:"data" pg:",notnull"`
}

type ConfigSchema struct {
	//nolint
	tableName string `pg:"?db_schema.config_schemas" json:"-"`
	Id        string `json:"id"`
	Version   string
	ModuleId  string
	Schema    schema.Schema
	CreatedAt time.Time
	UpdatedAt time.Time
}
