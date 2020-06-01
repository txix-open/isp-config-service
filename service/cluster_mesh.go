package service

import (
	"github.com/integration-system/isp-lib/v2/structure"
	"isp-config-service/store/state"
)

var (
	ClusterMesh clusterMeshService
)

type clusterMeshService struct{}

func (clusterMeshService) HandleUpdateBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.WritableState) {
	changed := state.WritableMesh().UpsertBackend(declaration)
	if changed {
		Discovery.BroadcastModuleAddresses(declaration.ModuleName, state.Mesh())
		Routes.BroadcastRoutes(state.Mesh())
	}
}

func (clusterMeshService) HandleDeleteBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.WritableState) {
	deleted := state.WritableMesh().DeleteBackend(declaration)
	if deleted {
		Discovery.BroadcastModuleAddresses(declaration.ModuleName, state.Mesh())
		Routes.BroadcastRoutes(state.Mesh())
	}
}
