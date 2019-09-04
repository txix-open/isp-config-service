package service

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/cluster"
	"isp-config-service/store/state"
)

var ApplyLogService applyLogService

type applyLogService struct{}

func (l applyLogService) HandleBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	state.UpsertBackend(declaration)
	logger.Debug("HandleBackendDeclarationCommand: upserted backend", declaration.ModuleName)
	DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
	DiscoveryService.BroadcastRoutes(state)
	return state, nil
}

func (l applyLogService) PrepareBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	var buf bytes.Buffer
	buf2 := make([]byte, 8)
	binary.BigEndian.PutUint64(buf2, cluster.BackendDeclarationCommand)
	buf.Write(buf2)
	err := json.NewEncoder(&buf).Encode(backend)
	if err != nil {
		logger.Fatal("PrepareBackendDeclarationCommand json encoding", err)
	}
	return buf.Bytes()
}
