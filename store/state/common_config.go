package state

import "isp-config-service/entity"

type WriteableCommonConfigStore interface {
	ReadonlyCommonConfigStore
	Upsert(config entity.CommonConfig)
	Delete(config entity.CommonConfig)
}

type ReadonlyCommonConfigStore interface {
	GetByIds(ids []string) []entity.CommonConfig
}

type CommonConfigStore struct {
	configs []entity.CommonConfig
}

func (cs CommonConfigStore) GetByIds(ids []string) []entity.CommonConfig {
	idsMap := StringsToMap(ids)
	configs := make([]entity.CommonConfig, 0, len(ids))
	for _, config := range cs.configs {
		if _, found := idsMap[config.Id]; found {
			configs = append(configs, config)
		}
	}
	return configs
}

func (cs *CommonConfigStore) Upsert(config entity.CommonConfig) {
	// TODO
}

func (cs *CommonConfigStore) Delete(config entity.CommonConfig) {
	// TODO
}

func NewCommonConfigStore() *CommonConfigStore {
	return &CommonConfigStore{configs: make([]entity.CommonConfig, 0)}
}
