package subs

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/integration-system/isp-lib/bootstrap"
	schema2 "github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/entity"
	"isp-config-service/service"
	"isp-config-service/store/state"
	"isp-config-service/ws"
)

func (h *socketEventHandler) handleModuleReady(conn ws.Conn, data []byte) string {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		return err.Error()
	}

	_, err = govalidator.ValidateStruct(declaration)
	if err != nil {
		return err.Error()
	}
	conn.SetBackendDeclaration(declaration)
	var changed bool
	h.store.VisitReadState(func(state state.ReadState) {
		changed = state.CheckBackendChanged(declaration)
	})
	if changed {
		command := cluster.PrepareUpdateBackendDeclarationCommand(declaration)
		applyLogResponse, err := h.cluster.SyncApply(command)
		if err != nil {
			return err.Error()
		}
		if applyLogResponse.ApplyError != "" {
			log.WithMetadata(map[string]interface{}{
				"comment":     applyLogResponse.Comment,
				"applyError":  applyLogResponse.ApplyError,
				"commandName": "UpdateBackendDeclarationCommand",
			}).Warn(codes.SyncApplyError, "apply command")
			return applyLogResponse.ApplyError
		}
	}
	return Ok
}

func (h *socketEventHandler) handleModuleRequirements(conn ws.Conn, data []byte) string {
	moduleName, err := conn.Parameters()
	log.Debugf(0, "handleModuleRequirements moduleName: %s", moduleName) // REMOVE
	if err != nil {
		return err.Error()
	}

	declaration := bootstrap.ModuleRequirements{}
	err = json.Unmarshal(data, &declaration)
	if err != nil {
		return err.Error()
	}

	h.store.VisitReadState(func(state state.ReadState) {
		if declaration.RequireRoutes {
			service.RoutesService.SubscribeRoutes(conn, state)
		}
		service.DiscoveryService.Subscribe(conn, declaration.RequiredModules, state)
	})
	return Ok
}

func (h *socketEventHandler) handleConfigSchema(conn ws.Conn, data []byte) string {
	moduleName, err := conn.Parameters()
	log.Debugf(0, "handleModuleRequirements moduleName: %s", moduleName) // REMOVE
	if err != nil {
		return err.Error()
	}

	configSchema := schema2.ConfigSchema{}
	if err := jsoniter.Unmarshal(data, &configSchema); err != nil {
		return err.Error()
	}
	module := new(entity.Module)
	h.store.VisitReadState(func(readState state.ReadState) {
		module = readState.GetModuleByName(moduleName)
	})
	if module == nil {
		return fmt.Sprintf("module with name %s not found", moduleName)
	}

	schema := entity.ConfigSchema{
		ModuleId: module.Id,
		Version:  configSchema.Version,
		Schema:   configSchema.Schema,
	}
	command := cluster.PrepareUpdateConfigSchemaCommand(schema)
	i, err := h.cluster.SyncApply(command)
	if err != nil {
		log.WithMetadata(map[string]interface{}{
			"answer": i,
		}).Warnf(codes.SyncApplyError, "apply ModuleSendConfigSchemaCommand %v", err)
	}
	return Ok
}

func (h *socketEventHandler) handleRequestConfig(conn ws.Conn, data []byte) {

}
