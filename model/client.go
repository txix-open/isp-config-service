package model

import (
	"github.com/integration-system/isp-lib/v2/database"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
)

var (
	DbClient = database.NewRxDbClient(
		database.WithSchemaEnsuring(),
		database.WithSchemaAutoInjecting(),
		database.WithMigrationsEnsuring(),
		database.WithInitializingErrorHandler(func(err *database.ErrorEvent) {
			log.Fatal(stdcodes.InitializingDbError, err.Error())
		}),
	)
	ConfigRep       ConfigRepository        = &configRepPg{rxClient: DbClient}
	CommonConfigRep CommonConfigRepository  = &commonConfigRepPg{rxClient: DbClient}
	SchemaRep       SchemaRepository        = &schemaRepPg{rxClient: DbClient}
	ModuleRep       ModulesRepository       = &modulesRepPg{rxClient: DbClient}
	VersionStoreRep VersionConfigRepository = &versionConfigRepPg{rxClient: DbClient}
)
