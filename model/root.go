package model

import (
	"github.com/go-pg/pg"
)

var (
	schema      = ""
	ConfigRep   ConfigRepository
	ModulesRep  ModulesRepository
	InstanceRep InstanceRepository
	SchemaRep   SchemaRepository
)

func SetSchema(s string) {
	schema = s
}

func InitDbManager(db *pg.DB) {
	ConfigRep = ConfigRepository{db}
	ModulesRep = ModulesRepository{db}
	InstanceRep = InstanceRepository{db}
	SchemaRep = SchemaRepository{db}
}
