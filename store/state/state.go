package state

import (
	"isp-config-service/entity"
)

type State struct {
	MeshStore          *Mesh
	ConfigsStore       *ConfigStore
	SchemasStore       *SchemaStore
	ModulesStore       *ModuleStore
	CommonConfigsStore *CommonConfigStore
}

func (s State) Mesh() ReadonlyMesh {
	return s.MeshStore
}

func (s State) Configs() ReadonlyConfigStore {
	return s.ConfigsStore
}

func (s State) Schemas() ReadonlySchemaStore {
	return s.SchemasStore
}

func (s State) Modules() ReadonlyModuleStore {
	return s.ModulesStore
}

func (s State) CommonConfigs() ReadonlyCommonConfigStore {
	return s.CommonConfigsStore
}

func (s *State) WritableMesh() WriteableMesh {
	return s.MeshStore
}

func (s *State) WritableConfigs() WriteableConfigStore {
	return s.ConfigsStore
}

func (s *State) WritableSchemas() WriteableSchemaStore {
	return s.SchemasStore
}

func (s *State) WritableModules() WriteableModuleStore {
	return s.ModulesStore
}

func (s *State) WritableCommonConfigs() WriteableCommonConfigStore {
	return s.CommonConfigsStore
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

func NewState() *State {
	return &State{
		MeshStore:          NewMesh(),
		ConfigsStore:       NewConfigStore(),
		SchemasStore:       NewSchemaStore(),
		ModulesStore:       NewModuleStore(),
		CommonConfigsStore: NewCommonConfigStore(),
	}
}

func NewStateFromSnapshot(configs []entity.Config, schemas []entity.ConfigSchema, modules []entity.Module, commConfigs []entity.CommonConfig) *State {
	return &State{
		MeshStore:          NewMesh(),
		ConfigsStore:       &ConfigStore{Configs: configs},
		SchemasStore:       &SchemaStore{Schemas: schemas},
		ModulesStore:       &ModuleStore{Modules: modules},
		CommonConfigsStore: &CommonConfigStore{Configs: commConfigs},
	}
}
