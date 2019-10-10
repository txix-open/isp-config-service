package state

import "isp-config-service/entity"

type ModuleStore struct {
	modules []entity.Module
}

func (ms *ModuleStore) Upsert(module entity.Module) {
	// TODO
}

func (ms *ModuleStore) GetByName(name string) (*entity.Module, error) {
	// TODO
	return nil, nil
}

func (ms *ModuleStore) GetById(id int32) (*entity.Module, error) {
	// TODO
	return nil, nil
}

func (ms *ModuleStore) GetActiveModules() ([]entity.Module, error) {
	// TODO
	return nil, nil
}

func NewModuleStore() *ModuleStore {
	return &ModuleStore{modules: make([]entity.Module, 0)}
}
