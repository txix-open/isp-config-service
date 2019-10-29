package subs

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	etp "github.com/integration-system/isp-etp-go"
	"github.com/integration-system/isp-lib/bootstrap"
	schema2 "github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/entity"
	"isp-config-service/service"
	"isp-config-service/store/state"
)

func (h *socketEventHandler) handleModuleReady(conn etp.Conn, data []byte) []byte {
	moduleName, _ := Parameters(conn)                             // REMOVE
	log.Debugf(0, "handleModuleReady moduleName: %s", moduleName) // REMOVE
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		return []byte(err.Error())
	}

	_, err = govalidator.ValidateStruct(declaration)
	if err != nil {
		return []byte(err.Error())
	}
	SetBackendDeclaration(conn, declaration)
	var changed bool
	h.store.VisitReadonlyState(func(state state.ReadonlyState) {
		changed = state.Mesh().CheckBackendChanged(declaration)
	})
	if changed {
		command := cluster.PrepareUpdateBackendDeclarationCommand(declaration)
		_, err = SyncApplyCommand(command, "UpdateBackendDeclarationCommand")
		if err != nil {
			return []byte(err.Error())
		}
	}
	return []byte(Ok)
}

func (h *socketEventHandler) handleModuleRequirements(conn etp.Conn, data []byte) []byte {
	moduleName, err := Parameters(conn)
	log.Debugf(0, "handleModuleRequirements moduleName: %s", moduleName) // REMOVE
	if err != nil {
		return []byte(err.Error())
	}

	declaration := bootstrap.ModuleRequirements{}
	err = json.Unmarshal(data, &declaration)
	if err != nil {
		return []byte(err.Error())
	}

	h.store.VisitReadonlyState(func(state state.ReadonlyState) {
		service.DiscoveryService.Subscribe(conn, declaration.RequiredModules, state.Mesh())
		if declaration.RequireRoutes {
			service.RoutesService.SubscribeRoutes(conn, state.Mesh())
		}
	})
	return []byte(Ok)
}

func (h *socketEventHandler) handleConfigSchema(conn etp.Conn, data []byte) []byte {
	moduleName, err := Parameters(conn)
	log.Debugf(0, "handleConfigSchema moduleName: %s", moduleName) // REMOVE
	if err != nil {
		return []byte(err.Error())
	}

	configSchema := schema2.ConfigSchema{}
	if err := json.Unmarshal(data, &configSchema); err != nil {
		return []byte(err.Error())
	}
	// TODO Костыль. Дважды посылаем ModuleConnected, т.к с момента первой отправки в handleConnect,
	//  состояние к конкретной ноде кластера может не успеть примениться
	now := state.GenerateDate()
	newModule := entity.Module{
		Id:              state.GenerateId(),
		Name:            moduleName,
		CreatedAt:       now,
		LastConnectedAt: now,
	}
	command := cluster.PrepareModuleConnectedCommand(newModule)
	_, err = SyncApplyCommand(command, "ModuleConnectedCommand")
	if err != nil {
		return []byte(err.Error())
	}
	// />

	module := new(entity.Module)
	h.store.VisitReadonlyState(func(readState state.ReadonlyState) {
		module = readState.Modules().GetByName(moduleName)
	})
	if module == nil {
		return []byte(fmt.Sprintf("module with name %s not found", moduleName))
	}

	schema := entity.ConfigSchema{
		Id:        state.GenerateId(),
		Version:   configSchema.Version,
		ModuleId:  module.Id,
		Schema:    configSchema.Schema,
		CreatedAt: now,
		UpdatedAt: now,
	}
	command = cluster.PrepareUpdateConfigSchemaCommand(schema)
	_, _ = SyncApplyCommand(command, "UpdateConfigSchemaCommand")

	var configs []entity.Config
	h.store.VisitReadonlyState(func(readState state.ReadonlyState) {
		configs = readState.Configs().GetByModuleIds([]string{module.Id})
	})
	if len(configs) == 0 {
		config := entity.Config{
			Id:        state.GenerateId(),
			Name:      module.Name,
			ModuleId:  module.Id,
			Data:      configSchema.DefaultConfig,
			CreatedAt: now,
			UpdatedAt: now,
		}
		upsertConfig := cluster.UpsertConfig{
			Config: config,
			Create: true,
		}

		command := cluster.PrepareUpsertConfigCommand(upsertConfig)
		_, _ = SyncApplyCommand(command, "UpsertConfigCommand")
	}
	return []byte(Ok)
}

func (h *socketEventHandler) handleRequestConfig(conn etp.Conn, data []byte) []byte {
	panic("not implemented")
}
