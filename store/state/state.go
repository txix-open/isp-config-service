package state

import (
	"isp-config-service/entity"
)

type State struct {
	mesh          *Mesh
	configs       *ConfigStore
	schemas       *SchemaStore
	modules       *ModuleStore
	commonConfigs *CommonConfigStore
}

func (s State) Mesh() ReadonlyMesh {
	return s.mesh
}

func (s State) Configs() ReadonlyConfigStore {
	return s.configs
}

func (s State) Schemas() ReadonlySchemaStore {
	return s.schemas
}

func (s State) Modules() ReadonlyModuleStore {
	return s.modules
}

func (s State) CommonConfigs() ReadonlyCommonConfigStore {
	return s.commonConfigs
}

func (s *State) WritableMesh() WriteableMesh {
	return s.mesh
}

func (s *State) WritableConfigs() WriteableConfigStore {
	return s.configs
}

func (s *State) WritableSchemas() WriteableSchemaStore {
	return s.schemas
}

func (s *State) WritableModules() WriteableModuleStore {
	return s.modules
}

func (s *State) WritableCommonConfigs() WriteableCommonConfigStore {
	return s.commonConfigs
}

type WritableState interface {
	ReadonlyState
	WritableMesh() WriteableMesh
	WritableConfigs() WriteableConfigStore
	WritableSchemas() WriteableSchemaStore
	WritableModules() WriteableModuleStore
	WritableCommonConfigs() WriteableCommonConfigStore
}

type ReadonlyState interface {
	Mesh() ReadonlyMesh
	Configs() ReadonlyConfigStore
	Schemas() ReadonlySchemaStore
	Modules() ReadonlyModuleStore
	CommonConfigs() ReadonlyCommonConfigStore
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
