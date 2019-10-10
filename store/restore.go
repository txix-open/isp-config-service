package store

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

func NewStateFromRepository() state.State {
	configs, err := model.ConfigRep.Snapshot()
	if err != nil {
		log.Fatalf(codes.RestoreFromRepositoryError, "config repository: %s", err)
	}
	schemas, err := model.SchemaRep.Snapshot()
	if err != nil {
		log.Fatalf(codes.RestoreFromRepositoryError, "schema repository: %s", err)
	}
	modules, err := model.ModuleRep.Snapshot()
	if err != nil {
		log.Fatalf(codes.RestoreFromRepositoryError, "module repository: %s", err)
	}
	commonConfigs, err := model.CommonConfigRep.Snapshot()
	if err != nil {
		log.Fatalf(codes.RestoreFromRepositoryError, "module repository: %s", err)
	}

	st := state.NewStateFromSnapshot(configs, schemas, modules, commonConfigs)
	return st
}
