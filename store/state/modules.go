package state

import (
	"isp-config-service/entity"
)

type WriteableModuleStore interface {
	ReadonlyModuleStore
	UpdateByName(module entity.Module)
	Create(entity.Module)
	DeleteByIds(ids []string) (deleted []entity.Module)
}

type ReadonlyModuleStore interface {
	GetByName(name string) *entity.Module
	GetById(id string) *entity.Module
	GetAll() []entity.Module
}

type ModuleStore struct {
	modules []entity.Module
}

func (ms ModuleStore) GetByName(name string) *entity.Module {
	for _, module := range ms.modules {
		if module.Name == name {
			return &module
		}
	}
	return nil
}

func (ms ModuleStore) GetById(id string) *entity.Module {
	for _, module := range ms.modules {
		if module.Id == id {
			return &module
		}
	}
	return nil
}

func (ms *ModuleStore) UpdateByName(module entity.Module) {
	for i := range ms.modules {
		if ms.modules[i].Name == module.Name {
			ms.modules[i] = module
			return
		}
	}
}

func (ms *ModuleStore) Create(module entity.Module) {
	ms.modules = append(ms.modules, module)
}

func (ms *ModuleStore) DeleteByIds(ids []string) []entity.Module {
	idsMap := StringsToMap(ids)
	var deleted []entity.Module
	for i := 0; i < len(ms.modules); i++ {
		id := ms.modules[i].Id
		if _, ok := idsMap[id]; ok {
			// change modules ordering
			deleted = append(deleted, ms.modules[i])
			ms.modules[i] = ms.modules[len(ms.modules)-1]
			ms.modules = ms.modules[:len(ms.modules)-1]
		}
	}
	return deleted
}

func (ms *ModuleStore) GetAll() []entity.Module {
	response := make([]entity.Module, len(ms.modules))
	copy(response, ms.modules)
	return response
}

func NewModuleStore() *ModuleStore {
	return &ModuleStore{modules: make([]entity.Module, 0)}
}
