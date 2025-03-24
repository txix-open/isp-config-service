package entity

import (
	"isp-config-service/entity/xtypes"
)

type Variable struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Value       string      `json:"value"`
	CreatedAt   xtypes.Time `json:"created_at"`
	UpdatedAt   xtypes.Time `json:"updated_at"`
}

type ConfigHasVariable struct {
	ConfigId     string `json:"config_id"`
	VariableName string `json:"variable_name"`
}
