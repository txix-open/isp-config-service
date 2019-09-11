package cluster

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
)

const (
	ApplyCommandEvent = "CONFIG_CLUSTER:APPLY_COMMAND"

	TokenParam   = "token"
	ClusterParam = "cluster"
)
const (
	_ = iota
	UpdateBackendDeclarationCommand
	DeleteBackendDeclarationCommand
)

func PrepareUpdateBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	return prepareCommand(UpdateBackendDeclarationCommand, backend)
}

func PrepareDeleteBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	return prepareCommand(DeleteBackendDeclarationCommand, backend)
}

func prepareCommand(command uint64, payload interface{}) []byte {
	cmd := make([]byte, 8)
	binary.BigEndian.PutUint64(cmd, command)
	buf := bytes.NewBuffer(cmd)
	err := json.NewEncoder(buf).Encode(payload)
	if err != nil {
		logger.Fatal("prepareCommand json encoding", err)
	}
	return buf.Bytes()
}
