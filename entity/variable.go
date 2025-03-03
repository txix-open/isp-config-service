package entity

import (
	"isp-config-service/entity/xtypes"
)

type Variable struct {
	Name        string
	Description string
	Type        string
	Value       string
	CreatedAt   xtypes.Time
	UpdatedAt   xtypes.Time
}

type ConfigHasVariable struct {
	ConfigId     string
	VariableName string
}
