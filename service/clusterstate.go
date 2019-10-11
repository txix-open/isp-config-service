package service

import (
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/store/state"
)

var (
	ClusterStateService clusterStateService
)

type clusterStateService struct{}

func (l clusterStateService) HandleUpdateBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	changed := state.UpsertBackend(declaration)
	if changed {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
		RoutesService.BroadcastRoutes(state)
	}
	return state, nil
}

func (l clusterStateService) HandleDeleteBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	deleted := state.DeleteBackend(declaration)
	if deleted {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
		RoutesService.BroadcastRoutes(state)
	}
	return state, nil
}
