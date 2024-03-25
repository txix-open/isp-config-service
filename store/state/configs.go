package state

import (
	"time"

	"isp-config-service/entity"
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
	GetByIds(ids []string) []entity.Config
	GetByModuleIds(ids []string) []entity.Config
	FilterByCommonConfigs(commonIds []string) []entity.Config
}

type ConfigStore struct {
	Configs []entity.Config
}

func (cs ConfigStore) GetActiveByModuleId(moduleId string) *entity.Config {
	for i := range cs.Configs {
		if cs.Configs[i].ModuleId == moduleId && cs.Configs[i].Active {
			conf := cs.Configs[i]
			return &conf
		}
	}
	return nil
}

func (cs ConfigStore) GetByIds(ids []string) []entity.Config {
	idsMap := StringsToMap(ids)
	result := make([]entity.Config, 0, len(ids))
	for i := range cs.Configs {
		if _, ok := idsMap[cs.Configs[i].Id]; ok {
			result = append(result, cs.Configs[i])
		}
	}
	return result
}

func (cs ConfigStore) GetByModuleIds(ids []string) []entity.Config {
	idsMap := StringsToMap(ids)
	result := make([]entity.Config, 0, len(ids))
	for i := range cs.Configs {
		if _, ok := idsMap[cs.Configs[i].ModuleId]; ok {
			result = append(result, cs.Configs[i])
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
	cs.Configs = append(cs.Configs, config)
	return config
}

func (cs *ConfigStore) UpdateById(config entity.Config) {
	for i := range cs.Configs {
		if cs.Configs[i].Id == config.Id {
			cs.Configs[i] = config
			break
		}
	}
}

func (cs *ConfigStore) calcNewVersion(moduleId string) int32 {
	var maxVersion int32 = 0
	for i := range cs.Configs {
		if cs.Configs[i].ModuleId == moduleId {
			if cs.Configs[i].Version > maxVersion {
				maxVersion = cs.Configs[i].Version
			}
		}
	}
	return maxVersion + 1
}

func (cs *ConfigStore) DeleteByIds(ids []string) int {
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
	for i := range cs.Configs {
		if cs.Configs[i].ModuleId == moduleId && cs.Configs[i].Active {
			cs.Configs[i].Active = false
			cs.Configs[i].UpdatedAt = date
			affected = append(affected, cs.Configs[i])
		}
	}
	return affected
}

func (cs *ConfigStore) FilterByCommonConfigs(commonIds []string) []entity.Config {
	idsMap := StringsToMap(commonIds)
	result := make([]entity.Config, 0)
	for i := range cs.Configs {
		for _, commonId := range cs.Configs[i].CommonConfigs {
			if _, ok := idsMap[commonId]; ok {
				result = append(result, cs.Configs[i])
				break
			}
		}
	}
	return result
}

func NewConfigStore() *ConfigStore {
	return &ConfigStore{Configs: make([]entity.Config, 0)}
}
