package cluster

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
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
		log.Fatalf(codes.PrepareLogCommandError, "prepare log command: %v", err)
	}
	return buf.Bytes()
}
