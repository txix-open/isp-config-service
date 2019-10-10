package store

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

func NewStateStoreFromRepository() *Store {
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

	store := &Store{
		state: state.NewStateFromSnapshot(configs, schemas, modules, commonConfigs),
	}

	store.handlers = map[uint64]func([]byte) error{
		cluster.UpdateBackendDeclarationCommand: store.applyUpdateBackendDeclarationCommand,
		cluster.DeleteBackendDeclarationCommand: store.applyDeleteBackendDeclarationCommand,
	}
	return store
}
