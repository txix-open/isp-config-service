package domain

import "time"

type Variable struct {
	Name              string
	Description       string
	Type              string
	Value             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ContainsInConfigs []LinkedConfig
}

type LinkedConfig struct {
	Id       string
	ModuleId string
	Name     string
}

type CreateVariableRequest struct {
	Name        string `validate:"required"`
	Description string
	Type        string `validate:"required"`
	Value       string `validate:"required"`
}

type UpdateVariableRequest struct {
	Name        string `validate:"required"`
	Description string
	Value       string `validate:"required"`
}

type UpsertVariableRequest struct {
	Name        string `validate:"required"`
	Description string `validate:"required"`
	Value       string `validate:"required"`
}

type VariableByNameRequest struct {
	Name string `validate:"required"`
}
