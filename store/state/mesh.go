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
	GetRoutes() structure.RoutingConfig
}

type Mesh struct {
	// { "ModuleName": { "address": BackendDeclaration, ...}, ... }
	modulesMap map[string]NodesMap
}

func (m Mesh) BackendExist(backend structure.BackendDeclaration) (exist bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.modulesMap[backend.ModuleName]; ok {
		if _, ok := nodes[address]; ok {
			exist = true
		}
	}
	return
}

func (m Mesh) GetModuleAddresses(moduleName string) []structure.AddressConfiguration {
	addressList := make([]structure.AddressConfiguration, 0)
	if nodes, ok := m.modulesMap[moduleName]; ok {
		for _, backend := range nodes {
			addressList = append(addressList, backend.Address)
		}
	}
	return addressList
}

func (m Mesh) GetRoutes() structure.RoutingConfig {
	routes := structure.RoutingConfig{}
	for _, node := range m.modulesMap {
		for _, backend := range node {
			routes = append(routes, backend)
		}
	}
	return routes
}

func (m Mesh) CheckBackendChanged(backend structure.BackendDeclaration) (changed bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.modulesMap[backend.ModuleName]; ok {
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
	if nodes, ok := m.modulesMap[backend.ModuleName]; ok {
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
		m.modulesMap[backend.ModuleName] = NodesMap{address: backend}
		changed = true
	}
	return
}

func (m *Mesh) DeleteBackend(backend structure.BackendDeclaration) (deleted bool) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.modulesMap[backend.ModuleName]; ok {
		if _, ok := nodes[address]; ok {
			delete(nodes, address)
			deleted = true
		}
		if len(nodes) == 0 {
			delete(m.modulesMap, backend.ModuleName)
		}
	}
	return
}

type NodesMap map[string]structure.BackendDeclaration

func NewMesh() *Mesh {
	return &Mesh{modulesMap: make(map[string]NodesMap)}
}
