package state

import "isp-config-service/entity"

type ConfigStore struct {
	configs []entity.Config
}

func (cs *ConfigStore) Upsert(config entity.Config) {
	// TODO
}

func (cs *ConfigStore) Delete(config entity.Config) {
	// TODO
}

func (cs *ConfigStore) Activate(config entity.Config) {
	// TODO
}

func (cs *ConfigStore) GetByModuleName(moduleName string) (*entity.Config, error) {
	// TODO
	return nil, nil
}

func (cs *ConfigStore) GetById(id int64) (*entity.Config, error) {
	// TODO
	return nil, nil
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{configs: make([]entity.Config, 0)}
}
