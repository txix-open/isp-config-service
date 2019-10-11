package state

import "isp-config-service/entity"

type CommonConfigStore struct {
	configs []entity.CommonConfig
}

func (cs *CommonConfigStore) Upsert(config entity.CommonConfig) {
	// TODO
}

func (cs *CommonConfigStore) Delete(config entity.CommonConfig) {
	// TODO
}

func (cs *CommonConfigStore) GetByIds(ids []string) []entity.CommonConfig {
	if len(ids) == 0 {
		return nil
	}
	idsMap := StringsToMap(ids)
	configs := make([]entity.CommonConfig, 0, len(ids))
	for _, config := range cs.configs {
		if _, found := idsMap[config.Id]; found {
			configs = append(configs, config)
		}
	}
	return configs
}

func NewCommonConfigStore() *CommonConfigStore {
	return &CommonConfigStore{configs: make([]entity.CommonConfig, 0)}
}
