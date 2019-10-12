package state

import (
	"github.com/integration-system/isp-lib/structure"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"time"
)

type State struct {
	mesh          *Mesh
	configs       *ConfigStore
	schemas       *SchemaStore
	modules       *ModuleStore
	commonConfigs *CommonConfigStore
}

type ReadState interface {
	CheckBackendChanged(backend structure.BackendDeclaration) (changed bool)
	GetModuleAddresses(moduleName string) []structure.AddressConfiguration
	GetRoutes() structure.RoutingConfig
	BackendExist(backend structure.BackendDeclaration) (exist bool)
	GetCompiledConfig(moduleName string) (map[string]interface{}, error)
	GetModuleByName(string) *entity.Module
}

func NewState() State {
	return State{
		mesh:          NewMesh(),
		configs:       NewConfigStore(),
		schemas:       NewSchemaStore(),
		modules:       NewModuleStore(),
		commonConfigs: NewCommonConfigStore(),
	}
}

func NewStateFromSnapshot(configs []entity.Config, schemas []entity.ConfigSchema, modules []entity.Module, commConfigs []entity.CommonConfig) State {
	return State{
		mesh:          NewMesh(),
		configs:       &ConfigStore{configs: configs},
		schemas:       &SchemaStore{schemas: schemas},
		modules:       &ModuleStore{modules: modules},
		commonConfigs: &CommonConfigStore{configs: commConfigs},
	}
}

func (s State) CheckBackendChanged(backend structure.BackendDeclaration) (changed bool) {
	return s.mesh.CheckBackendChanged(backend)
}

func (s *State) UpsertBackend(backend structure.BackendDeclaration) (changed bool) {
	return s.mesh.UpsertBackend(backend)
}

func (s State) GetModuleAddresses(moduleName string) []structure.AddressConfiguration {
	return s.mesh.GetModuleAddresses(moduleName)
}

func (s State) GetRoutes() structure.RoutingConfig {
	return s.mesh.GetRoutes()
}

func (s *State) DeleteBackend(backend structure.BackendDeclaration) (deleted bool) {
	return s.mesh.DeleteBackend(backend)
}

func (s State) BackendExist(backend structure.BackendDeclaration) (exist bool) {
	return s.mesh.BackendExist(backend)
}

func (s State) GetCompiledConfig(moduleName string) (map[string]interface{}, error) {
	module := s.modules.GetByName(moduleName)
	if module == nil {
		return nil, errors.Errorf("module with name %s not found", moduleName)
	}
	config := s.configs.GetActiveByModuleId(module.Id)
	if config == nil {
		return nil, errors.Errorf("no active configs for moduleId %s", module.Id)
	}
	commonConfigs := s.commonConfigs.GetByIds(config.CommonConfigs)
	configsToMerge := make([]map[string]interface{}, 0, len(commonConfigs))
	for _, common := range commonConfigs {
		configsToMerge = append(configsToMerge, common.Data)
	}
	configsToMerge = append(configsToMerge, config.Data)

	resultData := MergeNestedMaps(configsToMerge...)
	return resultData, nil
}

func (s *State) UpdateModuleLastConnected(moduleName string) entity.Module {
	existedModule := s.modules.GetByName(moduleName)
	if existedModule == nil {
		module := s.modules.Create(moduleName)
		return module
	}
	existedModule.LastConnectedAt = time.Now()
	s.modules.Update(*existedModule)
	return *existedModule
}

func (s *State) UpdateModuleLastDisconnected(moduleName string) entity.Module {
	existedModule := s.modules.GetByName(moduleName)
	if existedModule == nil {
		module := s.modules.Create(moduleName)
		return module
	}
	existedModule.LastDisconnectedAt = time.Now()
	s.modules.Update(*existedModule)
	return *existedModule
}

func (s State) GetModuleByName(moduleName string) *entity.Module {
	return s.modules.GetByName(moduleName)
}

func (s State) GetModuleById(moduleId string) *entity.Module {
	return s.modules.GetById(moduleId)
}

func (s *State) UpdateSchema(schema entity.ConfigSchema) entity.ConfigSchema {
	return s.schemas.Upsert(schema)
}
