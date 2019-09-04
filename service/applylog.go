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
	ClusterStoreService clusterStoreService
)

type clusterStoreService struct{}

func (l clusterStoreService) HandleBackendDeclarationCommand(declaration structure.BackendDeclaration, state state.State) (state.State, error) {
	state.UpsertBackend(declaration)
	logger.Debug("HandleBackendDeclarationCommand: upserted backend", declaration.ModuleName)
	DiscoveryService.BroadcastModuleAddresses(declaration.ModuleName, state)
	DiscoveryService.BroadcastRoutes(state)
	return state, nil
}

//TODO вынести как отдельную функция с парметрами вызова (int command, payload interface{}) ([]byte, error) для возможности переиспользования
func (l clusterStoreService) PrepareBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
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
