package store

import (
	"fmt"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/cluster"
	"isp-config-service/service"
)

func (s *Store) applyUpdateBackendDeclarationCommand(data []byte) error {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		return fmt.Errorf("UpdateBackendDeclarationCommand: %w", err)
	}
	state, err := service.ClusterStateService.HandleUpdateBackendDeclarationCommand(declaration, s.state)
	if err != nil {
		return fmt.Errorf("UpdateBackendDeclarationCommand: %w", err)
	}
	s.state = state
	return nil
}

func (s *Store) applyDeleteBackendDeclarationCommand(data []byte) error {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		return fmt.Errorf("DeleteBackendDeclarationCommand: %w", err)
	}
	state, err := service.ClusterStateService.HandleDeleteBackendDeclarationCommand(declaration, s.state)
	if err != nil {
		return err
	}
	s.state = state
	return nil
}

func (s *Store) applyModuleConnectedCommand(data []byte) error {
	moduleConnected := cluster.ModuleConnected{}
	err := json.Unmarshal(data, &moduleConnected)
	if err != nil {
		return fmt.Errorf("ModuleConnectedCommand: %w", err)
	}
	state, err := service.ClusterStateService.HandleModuleConnectedCommand(moduleConnected, s.state)
	if err != nil {
		return err
	}
	s.state = state
	return nil
}
