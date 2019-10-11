package state

import (
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/entity"
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
	GetModuleByName(string) (entity.Module, bool)
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

func (s State) GetModuleByName(moduleName string) (entity.Module, bool) {
	return s.modules.GetByName(moduleName)
}

func (s *State) GetModuleById(moduleId string) (entity.Module, bool) {
	return s.modules.GetById(moduleId)
}

func (s *State) UpdateSchema(schema entity.ConfigSchema) entity.ConfigSchema {
	return s.schemas.Upsert(schema)
}
