package state

import (
	"isp-config-service/entity"
)

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

func (cs *ConfigStore) GetActiveByModuleId(moduleId string) *entity.Config {
	for _, conf := range cs.configs {
		if conf.ModuleId == moduleId && conf.Active {
			return &conf
		}
	}
	return nil
}

func (cs *ConfigStore) GetById(id int64) (*entity.Config, error) {
	// TODO
	return nil, nil
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{configs: make([]entity.Config, 0)}
}
