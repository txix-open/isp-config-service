package store

import (
	"github.com/integration-system/isp-lib/structure"
	"github.com/pkg/errors"
	"isp-config-service/cluster"
	"isp-config-service/entity"
	"isp-config-service/service"
)

func (s *Store) applyUpdateBackendDeclarationCommand(data []byte) (interface{}, error) {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal structure.BackendDeclaration")
	}
	err = service.ClusterMeshService.HandleUpdateBackendDeclarationCommand(declaration, s.state)
	if err != nil {
		return nil, errors.WithMessage(err, "update backend declaration")
	}
	return nil, nil
}

func (s *Store) applyDeleteBackendDeclarationCommand(data []byte) (interface{}, error) {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal structure.BackendDeclaration")
	}
	err = service.ClusterMeshService.HandleDeleteBackendDeclarationCommand(declaration, s.state)
	if err != nil {
		return nil, errors.WithMessage(err, "delete backend declaration")
	}
	return nil, nil
}

func (s *Store) applyUpdateConfigSchema(data []byte) (interface{}, error) {
	schema := entity.ConfigSchema{}
	err := json.Unmarshal(data, &schema)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal entity.ConfigSchema")
	}
	err = service.SchemaService.HandleUpdateConfigSchema(schema, s.state)
	if err != nil {
		return nil, errors.WithMessage(err, "update config schema")
	}
	return nil, nil
}

func (s *Store) applyModuleConnectedCommand(data []byte) (interface{}, error) {
	moduleConnected := cluster.ModuleConnected{}
	err := json.Unmarshal(data, &moduleConnected)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal cluster.ModuleConnected")
	}
	err = service.ModuleRegistryService.HandleModuleConnectedCommand(moduleConnected, s.state)
	if err != nil {
		return nil, errors.WithMessage(err, "register module")
	}
	return nil, nil
}
