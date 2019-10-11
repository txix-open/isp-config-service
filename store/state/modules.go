package state

import (
	"isp-config-service/entity"
)

type ModuleStore struct {
	modules []entity.Module
}

func (ms *ModuleStore) Upsert(module entity.Module) {
	// TODO
}

func (ms *ModuleStore) GetByName(name string) (entity.Module, bool) {
	for _, module := range ms.modules {
		if module.Name == name {
			return module, true
		}
	}
	return entity.Module{}, false
}

func (ms *ModuleStore) GetById(id string) (entity.Module, bool) {
	for _, module := range ms.modules {
		if module.Id == id {
			return module, true
		}
	}
	return entity.Module{}, false
}

func (ms *ModuleStore) GetActiveModules() ([]entity.Module, error) {
	// TODO
	return nil, nil
}

func NewModuleStore() *ModuleStore {
	return &ModuleStore{modules: make([]entity.Module, 0)}
}
