package state

import (
	"fmt"
	"isp-config-service/entity"
	"time"
)

type ModuleStore struct {
	modules []entity.Module
}

func (ms *ModuleStore) Update(module entity.Module) {
	for i := range ms.modules {
		if ms.modules[i].Id == module.Id {
			ms.modules[i] = module
		}
	}
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

func (ms *ModuleStore) Create(name string) entity.Module {
	module := entity.Module{
		Id:                 GenerateId(),
		Name:               name,
		CreatedAt:          time.Now(),
		LastConnectedAt:    time.Now(),
		LastDisconnectedAt: time.Time{},
	}
	ms.modules = append(ms.modules, module)
	return module
}

func NewModuleStore() *ModuleStore {
	return &ModuleStore{modules: make([]entity.Module, 0)}
}
