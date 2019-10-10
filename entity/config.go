package entity

import (
	"encoding/json"
	"github.com/integration-system/isp-lib/config/schema"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"time"
)

type ConfigData map[string]interface{}

func (cd ConfigData) ToJSON() string {
	if bytes, err := json.Marshal(cd); err == nil {
		return string(bytes)
	} else {
		log.Warnf(codes.ConfigDataSerializeError, "could not serialize config data to JSON %v", err)
		return "{}"
	}
}

type Config struct {
	tableName     string     `sql:"?db_schema.configs" json:"-"`
	Id            int64      `json:"id"`
	Uuid          string     `json:"uuid" valid:"required~Required,uuid~must be a valid uuid"`
	Name          string     `json:"name" valid:"required~Required"`
	CommonConfigs []string   `json:"commonConfigIds" pg:",array"`
	Description   string     `json:"description"`
	ModuleId      int32      `json:"moduleId" valid:"required~Required"`
	Version       int32      `json:"version" sql:",null"`
	Active        bool       `json:"active" sql:",null"`
	CreatedAt     time.Time  `json:"createdAt" sql:",null"`
	UpdatedAt     time.Time  `json:"updatedAt" sql:",null"`
	Data          ConfigData `json:"data" sql:",notnull"`
}

type CommonConfig struct {
	tableName   string     `sql:"?db_schema.common_configs" json:"-"`
	Uuid        string     `json:"uuid" valid:"required~Required,uuid~must be a valid uuid"`
	Name        string     `json:"name" valid:"required~Required"`
	Description string     `json:"description"`
	Version     int32      `json:"version" sql:",null"`
	CreatedAt   time.Time  `json:"createdAt" sql:",null"`
	UpdatedAt   time.Time  `json:"updatedAt" sql:",null"`
	Data        ConfigData `json:"data" sql:",notnull"`
}

type ConfigSchema struct {
	tableName string `sql:"?db_schema.config_schemas" json:"-"`
	Id        int32
	Uuid      string `json:"uuid" valid:"required~Required,uuid~must be a valid uuid"`
	Version   string
	ModuleId  int32
	Schema    schema.Schema
	CreatedAt time.Time
	UpdatedAt time.Time
}
