package service

import (
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	ClusterMeshService clusterMeshService
)

type clusterMeshService struct{}

func (clusterMeshService) HandleUpdateBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.WritableState) error {
	changed := state.WritableMesh().UpsertBackend(declaration)
	if changed {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
		RoutesService.BroadcastRoutes(state)
	}
	return nil
}

func (clusterMeshService) HandleDeleteBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.WritableState) error {
	deleted := state.WritableMesh().DeleteBackend(declaration)
	if deleted {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
		RoutesService.BroadcastRoutes(state)
		module := ModuleRegistryService.updateModuleLastDisconnected(declaration.ModuleName, state.WritableModules())
		if holder.ClusterClient.IsLeader() {
			_, err := model.ModuleRep.Upsert(module)
			if err != nil {
				log.Errorf(codes.DatabaseOperationError, "upsert module: %v", err)
			}
		}
	}
	return nil
}
