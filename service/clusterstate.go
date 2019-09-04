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

var (
	ClusterStateService clusterStateService
)

type clusterStateService struct{}

func (l clusterStateService) HandleUpdateBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	state.UpsertBackend(declaration)
	logger.Debug("HandleUpdateBackendDeclarationCommand: upserted backend", declaration.ModuleName)
	DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
	RoutesService.BroadcastRoutes(state)
	return state, nil
}

func (l clusterStateService) HandleDeleteBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	state.DeleteBackend(declaration)
	logger.Debug("HandleDeleteBackendDeclarationCommand: deleted backend", declaration.ModuleName)
	DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
	RoutesService.BroadcastRoutes(state)
	return state, nil
}

func (l clusterStateService) PrepareUpdateBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	return l.prepareCommand(cluster.UpdateBackendDeclarationCommand, backend)
}

func (l clusterStateService) PrepareDeleteBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	return l.prepareCommand(cluster.DeleteBackendDeclarationCommand, backend)
}

func (l clusterStateService) prepareCommand(command uint64, payload interface{}) []byte {
	var buf bytes.Buffer
	buf2 := make([]byte, 8)
	binary.BigEndian.PutUint64(buf2, command)
	buf.Write(buf2)
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		logger.Fatal("prepareCommand json encoding", err)
	}
	return buf.Bytes()
}
