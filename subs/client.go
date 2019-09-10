package subs

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/cluster"
	"isp-config-service/service"
	"isp-config-service/store/state"
	"isp-config-service/ws"
)

func (h *socketEventHandler) handleModuleReady(conn ws.Conn, data []byte) string {
	declaration := structure.BackendDeclaration{}
	err := json.Unmarshal(data, &declaration)
	if err != nil {
		logger.Warnf("handleModuleDeclaration: %s, error parse json data: %s", conn.Id(), err.Error())
		return err.Error()
	}

	_, err = govalidator.ValidateStruct(declaration)
	if err != nil {
		errors := govalidator.ErrorsByField(err)
		logger.Warnf("SOCKET ROUTES ERROR, handleModuleDeclaration: %s, error validate routes data: %s", conn.Id(), errors)
		return err.Error()
	}
	conn.SetBackendDeclaration(declaration)
	var changed bool
	h.store.VisitReadState(func(state state.ReadState) {
		changed = state.CheckBackendChanged(declaration)
	})
	if changed {
		command := cluster.PrepareUpdateBackendDeclarationCommand(declaration)
		i, err := h.cluster.SyncApply(command)
		logger.Debug("cluster.SyncApply UpdateBackendDeclarationCommand:", i, err)
	}
	return Ok
}

func (h *socketEventHandler) handleModuleRequirements(conn ws.Conn, data []byte) string {
	logger.Debugf("onReceivedModuleRequirements: %s %s", conn.Id(), string(data))

	instanceUuid, moduleName, err := conn.Parameters()
	logger.Debug("instanceUuid:", instanceUuid, "moduleName:", moduleName) // REMOVE
	if err != nil {
		return err.Error()
	}

	declaration := bootstrap.ModuleRequirements{}
	err = json.Unmarshal(data, &declaration)
	if err != nil {
		logger.Debugf("onReceivedModuleRequirements: %s, error parse json data: %s", conn.Id(), err.Error())
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

func (h *socketEventHandler) handleConfigSchema(conn ws.Conn, data []byte) {

}

func (h *socketEventHandler) handleRequestConfig(conn ws.Conn, data []byte) {

}
