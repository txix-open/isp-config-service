package state

import (
	"github.com/integration-system/isp-lib/v2/config"
	"isp-config-service/conf"
	"isp-config-service/entity"
)

const DefaultVersionCount = 15

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
	limit := config.Get().(*conf.Configuration).VersionConfigCount
	if limit <= 0 {
		limit = DefaultVersionCount
	}
	var removedId string
	store, found := s.VersionByConfigId[req.ConfigId]
	if found {
		if len(store) >= limit {
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
		return s.reverse(response)
	}
	return []entity.VersionConfig{}
}

func (s VersionConfigStore) reverse(cfg []entity.VersionConfig) []entity.VersionConfig {
	for i, j := 0, len(cfg)-1; i < j; i, j = i+1, j-1 {
		cfg[i], cfg[j] = cfg[j], cfg[i]
	}
	return cfg
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
