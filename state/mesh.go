package state

import "github.com/integration-system/isp-lib/structure"

type Mesh struct {
	modulesMap map[string]NodesMap
}

func (m *Mesh) UpsertBackend(backend structure.BackendDeclaration) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.modulesMap[backend.ModuleName]; ok {
		nodes[address] = backend
	} else {
		m.modulesMap[backend.ModuleName] = NodesMap{address: backend}
	}
}

func (m *Mesh) DeleteBackend(backend structure.BackendDeclaration) {
	address := backend.Address.GetAddress()
	if nodes, ok := m.modulesMap[backend.ModuleName]; ok {
		delete(nodes, address)
		if len(nodes) == 0 {
			delete(m.modulesMap, backend.ModuleName)
		}
	}
}

type NodesMap map[string]structure.BackendDeclaration

func NewMesh() *Mesh {
	return &Mesh{modulesMap: make(map[string]NodesMap)}
}
