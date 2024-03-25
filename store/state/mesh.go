package state

import (
	"github.com/integration-system/isp-lib/v2/structure"
)

type WriteableMesh interface {
	ReadonlyMesh
	UpsertBackend(backend structure.BackendDeclaration) (changed bool)
	DeleteBackend(backend structure.BackendDeclaration) (deleted bool)
}

type ReadonlyMesh interface {
	GetModuleAddresses(moduleName string) []structure.AddressConfiguration
	GetBackends(module string) []structure.BackendDeclaration
	GetRoutes() structure.RoutingConfig
}

type NodesMap map[string]structure.BackendDeclaration

type Mesh struct {
	// { "ModuleName": { "address": BackendDeclaration, ...}, ... }
	ModulesMap map[string]NodesMap
}

func NewMesh() *Mesh {
	return &Mesh{ModulesMap: make(map[string]NodesMap)}
}

func (m Mesh) GetBackends(module string) []structure.BackendDeclaration {
	declarations := make([]structure.BackendDeclaration, 0)
	if nodes, ok := m.ModulesMap[module]; ok {
		for _, backend := range nodes {
			declarations = append(declarations, backend)
		}
	}
	return declarations
}

func (m Mesh) GetModuleAddresses(moduleName string) []structure.AddressConfiguration {
	addressList := make([]structure.AddressConfiguration, 0)
	if nodes, ok := m.ModulesMap[moduleName]; ok {
		for _, backend := range nodes {
			addressList = append(addressList, backend.Address)
		}
	}
	return addressList
}

func (m Mesh) GetRoutes() structure.RoutingConfig {
	routes := structure.RoutingConfig{}
	for _, nodes := range m.ModulesMap {
		for _, backend := range nodes {
			routes = append(routes, backend)
		}
	}
	return routes
}

func (m *Mesh) UpsertBackend(backend structure.BackendDeclaration) bool {
	address := backend.Address.GetAddress()

	nodes, ok := m.ModulesMap[backend.ModuleName]
	if !ok {
		m.ModulesMap[backend.ModuleName] = NodesMap{address: backend}
		return true
	}

	old, ok := nodes[address]
	if !ok {
		nodes[address] = backend
		return true
	}

	if old.Version != backend.Version || !old.IsPathsEqual(backend.Endpoints) || old.LibVersion != backend.LibVersion {
		nodes[address] = backend
		return true
	}

	return false
}

func (m *Mesh) DeleteBackend(backend structure.BackendDeclaration) (deleted bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.ModulesMap[backend.ModuleName]; ok {
		if _, ok := nodes[address]; ok {
			delete(nodes, address)
			deleted = true
		}
		if len(nodes) == 0 {
			delete(m.ModulesMap, backend.ModuleName)
		}
	}
	return
}
