package state

import (
	"github.com/integration-system/isp-lib/structure"
)

type State struct {
	mesh *Mesh
}

func NewState() State {
	return State{
		mesh: NewMesh(),
	}
}
func (s *State) CheckBackendChanged(backend structure.BackendDeclaration) (changed bool) {
	return s.mesh.CheckBackendChanged(backend)
}

func (s *State) UpsertBackend(backend structure.BackendDeclaration) (changed bool) {
	return s.mesh.UpsertBackend(backend)
}

func (s *State) GetModuleAddresses(moduleName string) []structure.AddressConfiguration {
	return s.mesh.GetModuleAddresses(moduleName)
}

func (s *State) GetRoutes() structure.RoutingConfig {
	return s.mesh.GetRoutes()
}
