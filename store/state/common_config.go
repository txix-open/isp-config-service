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

func (cs *CommonConfigStore) GetById(id int64) (*entity.CommonConfig, error) {
	// TODO
	return nil, nil
}

func NewCommonConfigStore() *CommonConfigStore {
	return &CommonConfigStore{configs: make([]entity.CommonConfig, 0)}
}
