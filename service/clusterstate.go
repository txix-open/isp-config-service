package service

import (
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	ClusterStateService clusterStateService
)

type clusterStateService struct{}

func (clusterStateService) HandleUpdateBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	changed := state.UpsertBackend(declaration)
	if changed {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
		RoutesService.BroadcastRoutes(state)
	}
	return state, nil
}

func (clusterStateService) HandleDeleteBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	deleted := state.DeleteBackend(declaration)
	if deleted {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
		RoutesService.BroadcastRoutes(state)
		module := state.UpdateModuleLastDisconnected(declaration.ModuleName)
		if holder.ClusterClient.IsLeader() {
			_, err := model.ModuleRep.Upsert(module)
			if err != nil {
				log.Errorf(codes.DatabaseOperationError, "upsert module: %v", err)
			}
		}
	}
	return state, nil
}

func (clusterStateService) HandleModuleConnectedCommand(moduleConnected cluster.ModuleConnected, state state.State) (state.State, error) {
	module := state.UpdateModuleLastConnected(moduleConnected.ModuleName)
	if holder.ClusterClient.IsLeader() {
		_, err := model.ModuleRep.Upsert(module)
		if err != nil {
			log.Errorf(codes.DatabaseOperationError, "upsert module: %v", err)
		}
	}
	return state, nil
}
