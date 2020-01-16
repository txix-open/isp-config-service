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
	Modules []entity.Module
}

func (ms ModuleStore) GetByName(name string) *entity.Module {
	for _, module := range ms.Modules {
		if module.Name == name {
			return &module
		}
	}
	return nil
}

func (ms ModuleStore) GetById(id string) *entity.Module {
	for _, module := range ms.Modules {
		if module.Id == id {
			return &module
		}
	}
	return nil
}

func (ms *ModuleStore) UpdateByName(module entity.Module) {
	for i := range ms.Modules {
		if ms.Modules[i].Name == module.Name {
			ms.Modules[i] = module
			return
		}
	}
}

func (ms *ModuleStore) Create(module entity.Module) {
	ms.Modules = append(ms.Modules, module)
}

func (ms *ModuleStore) DeleteByIds(ids []string) []entity.Module {
	idsMap := StringsToMap(ids)
	var deleted []entity.Module
	for i := 0; i < len(ms.Modules); i++ {
		id := ms.Modules[i].Id
		if _, ok := idsMap[id]; ok {
			// change modules ordering
			deleted = append(deleted, ms.Modules[i])
			ms.Modules[i] = ms.Modules[len(ms.Modules)-1]
			ms.Modules = ms.Modules[:len(ms.Modules)-1]
		}
	}
	return deleted
}

func (ms *ModuleStore) GetAll() []entity.Module {
	response := make([]entity.Module, len(ms.Modules))
	copy(response, ms.Modules)
	return response
}

func NewModuleStore() *ModuleStore {
	return &ModuleStore{Modules: make([]entity.Module, 0)}
}
