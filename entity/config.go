package entity

import (
	"encoding/json"
	"github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/logger"
	"time"
)

type ConfigData map[string]interface{}

func (cd ConfigData) ToJSON() string {
	if bytes, err := json.Marshal(cd); err == nil {
		return string(bytes)
	} else {
		logger.Warn("Could not serialize config data to JSON", err)
		return "{}"
	}
}

type Config struct {
	Id          int64      `json:"id"`
	Name        string     `json:"name" valid:"required~Required"`
	Description string     `json:"description"`
	ModuleId    int32      `json:"moduleId" valid:"required~Required"`
	Version     int32      `json:"version" sql:",null"`
	Active      bool       `json:"active" sql:",null"`
	CreatedAt   time.Time  `json:"createdAt" sql:",null"`
	UpdatedAt   time.Time  `json:"updatedAt" sql:",null"`
	Data        ConfigData `json:"data" sql:",notnull"` // Объект конфигурации
}

type ConfigSchema struct {
	Id        int32
	Version   string
	ModuleId  int32
	Schema    schema.Schema
	CreatedAt time.Time
	UpdatedAt time.Time
}
