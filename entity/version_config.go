package entity

type VersionConfig struct {
	//nolint
	tableName     string `pg:"?db_schema.version_config" json:"-"`
	Id            string
	ConfigVersion int32
	ConfigId      string
	Data          ConfigData `json:"data" pg:",notnull"`
}
