package state

import (
	"github.com/integration-system/isp-lib/structure"
)

type State struct {
	mesh *Mesh
}

type ReadState interface {
	CheckBackendChanged(backend structure.BackendDeclaration) (changed bool)
	GetModuleAddresses(moduleName string) []structure.AddressConfiguration
	GetRoutes() structure.RoutingConfig
	BackendExist(backend structure.BackendDeclaration) (exist bool)
}

func NewState() State {
	return State{
		mesh: NewMesh(),
	}
}

func (s State) CheckBackendChanged(backend structure.BackendDeclaration) (changed bool) {
	return s.mesh.CheckBackendChanged(backend)
}

func (s *State) UpsertBackend(backend structure.BackendDeclaration) (changed bool) {
	return s.mesh.UpsertBackend(backend)
}

func (s State) GetModuleAddresses(moduleName string) []structure.AddressConfiguration {
	return s.mesh.GetModuleAddresses(moduleName)
}

func (s State) GetRoutes() structure.RoutingConfig {
	return s.mesh.GetRoutes()
}

func (s *State) DeleteBackend(backend structure.BackendDeclaration) {
	s.mesh.DeleteBackend(backend)
}

func (s State) BackendExist(backend structure.BackendDeclaration) (exist bool) {
	return s.mesh.BackendExist(backend)
}
