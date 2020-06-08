package state

import (
	"github.com/integration-system/isp-lib/v2/config"
	"isp-config-service/conf"
	"isp-config-service/entity"
)

const defaultVersionCount = 5

type WriteableVersionConfigStore interface {
	ReadonlyVersionConfigStore
	Update(config entity.VersionConfig) (removedId string)
	Delete(id string)
}

type ReadonlyVersionConfigStore interface {
	GetByConfigId(id string) []entity.VersionConfig
}

type VersionConfigStore struct {
	VersionByConfigId map[string][]entity.VersionConfig
}

func (s *VersionConfigStore) Update(req entity.VersionConfig) string {
	count := config.Get().(*conf.Configuration).VersionConfigCount
	if count <= 0 {
		count = defaultVersionCount
	}
	var removedId string
	store, found := s.VersionByConfigId[req.ConfigId]
	if found {
		if len(store) >= count {
			removedId = store[0].Id
			store = append(store[1:], req)
		} else {
			store = append(store, req)
		}
	} else {
		store = []entity.VersionConfig{req}
	}
	s.VersionByConfigId[req.ConfigId] = store
	return removedId
}

func (s *VersionConfigStore) Delete(id string) {
	for cfgId, versionCfg := range s.VersionByConfigId {
		for i, versionConfig := range versionCfg {
			if versionConfig.Id == id {
				newVersionCfg := versionCfg[:i+copy(versionCfg[i:], versionCfg[i+1:])]
				if len(newVersionCfg) == 0 {
					delete(s.VersionByConfigId, cfgId)
				} else {
					s.VersionByConfigId[cfgId] = newVersionCfg
				}
				return
			}
		}
	}
}

func (s VersionConfigStore) GetByConfigId(id string) []entity.VersionConfig {
	response, found := s.VersionByConfigId[id]
	if found {
		return response
	}
	return []entity.VersionConfig{}
}

func NewVersionConfigStore() *VersionConfigStore {
	return &VersionConfigStore{
		VersionByConfigId: make(map[string][]entity.VersionConfig),
	}
}

func NewVersionConfigStoreFromSnapshot(req []entity.VersionConfig) *VersionConfigStore {
	store := NewVersionConfigStore()
	for i := 0; i < len(req); i++ {
		_ = store.Update(req[i])
	}
	return store
}
