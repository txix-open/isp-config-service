package cluster

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/entity"
)

const (
	ApplyCommandEvent = "CONFIG_CLUSTER:APPLY_COMMAND"
	CommandSizeBytes  = 8

	GETParam = "cluster"
)

type Command uint64

// needed: go get golang.org/x/tools/cmd/stringer
//go:generate stringer -type=Command -output const_string.go

const (
	_ Command = iota
	UpdateBackendDeclarationCommand
	DeleteBackendDeclarationCommand

	UpdateConfigSchemaCommand

	ModuleConnectedCommand
	ModuleDisconnectedCommand
	DeleteModulesCommand
	DeleteVersionConfigCommand

	ActivateConfigCommand
	DeleteConfigsCommand
	UpsertConfigCommand

	DeleteCommonConfigsCommand
	UpsertCommonConfigCommand

	BroadcastEventCommand
)

func ParseCommand(data []byte) Command {
	return Command(binary.BigEndian.Uint64(data[:CommandSizeBytes]))
}

func PrepareUpdateBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	return prepareCommand(UpdateBackendDeclarationCommand, backend)
}

func PrepareDeleteBackendDeclarationCommand(backend structure.BackendDeclaration) []byte {
	return prepareCommand(DeleteBackendDeclarationCommand, backend)
}

func PrepareModuleConnectedCommand(module entity.Module) []byte {
	return prepareCommand(ModuleConnectedCommand, module)
}

func PrepareModuleDisconnectedCommand(module entity.Module) []byte {
	return prepareCommand(ModuleDisconnectedCommand, module)
}

func PrepareDeleteModulesCommand(ids []string) []byte {
	return prepareCommand(DeleteModulesCommand, DeleteModules{Ids: ids})
}

func PrepareActivateConfigCommand(configID string, date time.Time) []byte {
	return prepareCommand(ActivateConfigCommand, ActivateConfig{ConfigId: configID, Date: date})
}

func PrepareDeleteConfigsCommand(ids []string) []byte {
	return prepareCommand(DeleteConfigsCommand, DeleteModules{Ids: ids})
}

func PrepareUpsertConfigCommand(config UpsertConfig) []byte {
	return prepareCommand(UpsertConfigCommand, config)
}

func PrepareUpdateConfigSchemaCommand(schema entity.ConfigSchema) []byte {
	return prepareCommand(UpdateConfigSchemaCommand, schema)
}

func PrepareDeleteCommonConfigsCommand(id string) []byte {
	return prepareCommand(DeleteCommonConfigsCommand, DeleteCommonConfig{Id: id})
}

func PrepareUpsertCommonConfigCommand(config UpsertCommonConfig) []byte {
	return prepareCommand(UpsertCommonConfigCommand, config)
}

func PrepareBroadcastEventCommand(event BroadcastEvent) []byte {
	return prepareCommand(BroadcastEventCommand, event)
}

func PrepareDeleteConfigVersionCommand(id string) []byte {
	return prepareCommand(DeleteVersionConfigCommand, Identity{Id: id})
}

func prepareCommand(command Command, payload interface{}) []byte {
	cmd := make([]byte, 8)
	binary.BigEndian.PutUint64(cmd, uint64(command))
	buf := bytes.NewBuffer(cmd)
	err := json.NewEncoder(buf).Encode(payload)
	if err != nil {
		log.Fatalf(codes.PrepareLogCommandError, "prepare log command: %v", err)
	}
	return buf.Bytes()
}
