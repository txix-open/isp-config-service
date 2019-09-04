package store

import (
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/service"
)

func (s *Store) applyBackendDeclarationCommand(data []byte) error {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		logger.Warnf("Store.applyBackendDeclarationCommand: %s, error parse json data: %s", err.Error())
		return err
	}
	state, err := service.ApplyLogService.HandleBackendDeclarationCommand(declaration, s.state)
	if err != nil {
		return err
	}
	s.state = state
	return nil
}
