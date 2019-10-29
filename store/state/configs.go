package state

import (
	"isp-config-service/entity"
	"time"
)

const (
	configDefaultName = "unnamed"
)

type WriteableConfigStore interface {
	ReadonlyConfigStore
	Create(config entity.Config) entity.Config
	UpdateById(config entity.Config)
	DeleteByIds(ids []string) int
	Activate(config entity.Config, date time.Time) (affected []entity.Config)
}

type ReadonlyConfigStore interface {
	GetActiveByModuleId(moduleId string) *entity.Config
	GetActiveByModuleName(moduleId string) *entity.Config
	GetByIds(ids []string) []entity.Config
	GetByModuleIds(ids []string) []entity.Config
	FilterByCommonConfigs(commonIds []string) []entity.Config
}

type ConfigStore struct {
	configs []entity.Config
}

func (cs ConfigStore) GetActiveByModuleId(moduleId string) *entity.Config {
	for _, conf := range cs.configs {
		if conf.ModuleId == moduleId && conf.Active {
			return &conf
		}
	}
	return nil
}

func (cs ConfigStore) GetByIds(ids []string) []entity.Config {
	idsMap := StringsToMap(ids)
	result := make([]entity.Config, 0, len(ids))
	for _, conf := range cs.configs {
		if _, ok := idsMap[conf.Id]; ok {
			result = append(result, conf)
		}
	}
	return result
}

func (cs ConfigStore) GetByModuleIds(ids []string) []entity.Config {
	idsMap := StringsToMap(ids)
	result := make([]entity.Config, 0, len(ids))
	for _, conf := range cs.configs {
		if _, ok := idsMap[conf.ModuleId]; ok {
			result = append(result, conf)
		}
	}
	return result
}

func (cs *ConfigStore) Create(config entity.Config) entity.Config {
	config.Version = cs.calcNewVersion(config.ModuleId)
	if config.Version == 1 {
		config.Active = true
	}
	if config.Name == "" {
		config.Name = configDefaultName
	}
	if config.Data == nil {
		config.Data = make(entity.ConfigData)
	}
	if config.CommonConfigs == nil {
		config.CommonConfigs = make([]string, 0)
	}
	cs.configs = append(cs.configs, config)
	return config
}

func (cs *ConfigStore) UpdateById(config entity.Config) {
	for i := range cs.configs {
		if cs.configs[i].Id == config.Id {
			cs.configs[i] = config
			break
		}
	}
}

func (cs *ConfigStore) calcNewVersion(moduleId string) int32 {
	var maxVersion int32 = 0
	for i := range cs.configs {
		if cs.configs[i].ModuleId == moduleId {
			if cs.configs[i].Version > maxVersion {
				maxVersion = cs.configs[i].Version
			}
		}
	}
	return maxVersion + 1
}

func (cs *ConfigStore) DeleteByIds(ids []string) int {
	idsMap := StringsToMap(ids)
	var deleted int
	for i := 0; i < len(cs.configs); i++ {
		id := cs.configs[i].Id
		if _, ok := idsMap[id]; ok {
			// change configs ordering
			cs.configs[i] = cs.configs[len(cs.configs)-1]
			cs.configs = cs.configs[:len(cs.configs)-1]
			deleted++
		}
	}
	return deleted
}

func (cs *ConfigStore) Activate(config entity.Config, date time.Time) []entity.Config {
	affected := cs.deactivate(config.ModuleId, date)
	config.Active = true
	config.UpdatedAt = date
	cs.UpdateById(config)
	affected = append(affected, config)
	return affected
}

func (cs *ConfigStore) deactivate(moduleId string, date time.Time) []entity.Config {
	affected := make([]entity.Config, 0)
	for i := range cs.configs {
		if cs.configs[i].ModuleId == moduleId && cs.configs[i].Active {
			cs.configs[i].Active = false
			cs.configs[i].UpdatedAt = date
			affected = append(affected, cs.configs[i])
		}
	}
	return affected
}

func (cs *ConfigStore) FilterByCommonConfigs(commonIds []string) []entity.Config {
	idsMap := StringsToMap(commonIds)
	result := make([]entity.Config, 0)
	for _, cfg := range cs.configs {
		for _, commonId := range cfg.CommonConfigs {
			if _, ok := idsMap[commonId]; ok {
				result = append(result, cfg)
				break
			}
		}
	}
	return result
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{configs: make([]entity.Config, 0)}
}
