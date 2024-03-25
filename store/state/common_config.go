package state

import (
	"isp-config-service/entity"
)

type WriteableCommonConfigStore interface {
	ReadonlyCommonConfigStore
	Create(config entity.CommonConfig) entity.CommonConfig
	UpdateById(config entity.CommonConfig)
	DeleteByIds(ids []string) int
}

type ReadonlyCommonConfigStore interface {
	GetByIds(ids []string) []entity.CommonConfig
	GetByName(name string) []entity.CommonConfig
	GetAll() []entity.CommonConfig
}

type CommonConfigStore struct {
	Configs []entity.CommonConfig
}

func (cs CommonConfigStore) GetByIds(ids []string) []entity.CommonConfig {
	idsMap := StringsToMap(ids)
	configs := make([]entity.CommonConfig, 0, len(ids))
	for _, config := range cs.Configs {
		if _, found := idsMap[config.Id]; found {
			configs = append(configs, config)
		}
	}
	return configs
}

func (cs CommonConfigStore) GetByName(name string) []entity.CommonConfig {
	configs := make([]entity.CommonConfig, 0)
	for _, config := range cs.Configs {
		if config.Name == name {
			configs = append(configs, config)
		}
	}
	return configs
}

func (cs *CommonConfigStore) Create(config entity.CommonConfig) entity.CommonConfig {
	if config.Name == "" {
		config.Name = configDefaultName
	}
	if config.Data == nil {
		config.Data = make(entity.ConfigData)
	}
	cs.Configs = append(cs.Configs, config)
	return config
}

func (cs *CommonConfigStore) UpdateById(config entity.CommonConfig) {
	for i := range cs.Configs {
		if cs.Configs[i].Id == config.Id {
			cs.Configs[i] = config
			break
		}
	}
}

func (cs *CommonConfigStore) DeleteByIds(ids []string) int {
	idsMap := StringsToMap(ids)
	var deleted int
	for i := 0; i < len(cs.Configs); i++ {
		id := cs.Configs[i].Id
		if _, ok := idsMap[id]; ok {
			// change configs ordering
			cs.Configs[i] = cs.Configs[len(cs.Configs)-1]
			cs.Configs = cs.Configs[:len(cs.Configs)-1]
			deleted++
		}
	}
	return deleted
}

func (cs *CommonConfigStore) GetAll() []entity.CommonConfig {
	response := make([]entity.CommonConfig, len(cs.Configs))
	copy(response, cs.Configs)
	return response
}

func NewCommonConfigStore() *CommonConfigStore {
	return &CommonConfigStore{Configs: make([]entity.CommonConfig, 0)}
}
