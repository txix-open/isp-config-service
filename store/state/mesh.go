package state

import (
	"github.com/integration-system/isp-lib/structure"
)

type WriteableMesh interface {
	ReadonlyMesh
	UpsertBackend(backend structure.BackendDeclaration) (changed bool)
	DeleteBackend(backend structure.BackendDeclaration) (deleted bool)
}

type ReadonlyMesh interface {
	CheckBackendChanged(backend structure.BackendDeclaration) (changed bool)
	BackendExist(backend structure.BackendDeclaration) (exist bool)
	GetModuleAddresses(moduleName string) []structure.AddressConfiguration
	GetBackends(module string) []structure.BackendDeclaration
	GetRoutes() structure.RoutingConfig
}

type NodesMap map[string]structure.BackendDeclaration

type Mesh struct {
	// { "ModuleName": { "address": BackendDeclaration, ...}, ... }
	ModulesMap map[string]NodesMap
}

func (m Mesh) BackendExist(backend structure.BackendDeclaration) (exist bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.ModulesMap[backend.ModuleName]; ok {
		if _, ok := nodes[address]; ok {
			exist = true
		}
	}
	return
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
	for _, node := range m.ModulesMap {
		for _, backend := range node {
			routes = append(routes, backend)
		}
	}
	return routes
}

func (m Mesh) CheckBackendChanged(backend structure.BackendDeclaration) (changed bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.ModulesMap[backend.ModuleName]; ok {
		if oldBackend, ok := nodes[address]; ok {
			if oldBackend.Version != backend.Version || !oldBackend.IsPathsEqual(backend.Endpoints) || oldBackend.LibVersion != backend.LibVersion {
				changed = true
			}
		} else {
			changed = true
		}
	} else {
		changed = true
	}
	return
}

func (m *Mesh) UpsertBackend(backend structure.BackendDeclaration) (changed bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.ModulesMap[backend.ModuleName]; ok {
		if oldBackend, ok := nodes[address]; ok {
			if oldBackend.Version != backend.Version || !oldBackend.IsPathsEqual(backend.Endpoints) || oldBackend.LibVersion != backend.LibVersion {
				nodes[address] = backend
				changed = true
			}
		} else {
			nodes[address] = backend
			changed = true
		}
	} else {
		m.ModulesMap[backend.ModuleName] = NodesMap{address: backend}
		changed = true
	}
	return
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

func NewMesh() *Mesh {
	return &Mesh{ModulesMap: make(map[string]NodesMap)}
}
