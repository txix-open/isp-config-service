package entity

import (
	"time"
)

type VersionConfig struct {
	//nolint
	tableName     string `pg:"?db_schema.version_config" json:"-"`
	Id            string
	ConfigId      string
	ConfigVersion int32
	Data          ConfigData `json:"data,omitempty" pg:",notnull"`
	CreatedAt     time.Time
}
