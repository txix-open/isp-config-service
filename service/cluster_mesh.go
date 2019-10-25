package service

import (
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/store/state"
)

var (
	ClusterMeshService clusterMeshService
)

type clusterMeshService struct{}

func (clusterMeshService) HandleUpdateBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.WritableState) {
	changed := state.WritableMesh().UpsertBackend(declaration)
	if changed {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state.Mesh())
		RoutesService.BroadcastRoutes(state.Mesh())
	}
}

func (clusterMeshService) HandleDeleteBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.WritableState) {
	deleted := state.WritableMesh().DeleteBackend(declaration)
	if deleted {
		DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state.Mesh())
		RoutesService.BroadcastRoutes(state.Mesh())
	}
}
