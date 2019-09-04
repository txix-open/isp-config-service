package subs

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"isp-config-service/service"
	"isp-config-service/ws"
)

func (h *socketEventHandler) handleModuleReady(conn ws.Conn, data []byte) {
	instanceUuid, moduleName, err := conn.Parameters()
	logger.Debug("instanceUuid:", instanceUuid, "moduleName:", moduleName) // REMOVE
	if err != nil {
		_ = conn.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		return
	}

	declaration := structure.BackendDeclaration{}
	err = json.Unmarshal(data, &declaration)
	if err != nil {
		_ = conn.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		logger.Warnf("handleModuleDeclaration: %s, error parse json data: %s", conn.Id(), err.Error())
		return
	}

	_, err = govalidator.ValidateStruct(declaration)
	if err != nil {
		errors := govalidator.ErrorsByField(err)
		_ = conn.Emit(utils.ErrorConnection, map[string]map[string]string{"error": errors})
		logger.Warnf("SOCKET ROUTES ERROR, handleModuleDeclaration: %s, error validate routes data: %s", conn.Id(), errors)
	} else if h.store.GetState().CheckBackendChanged(declaration) {
		// TODO h.store.GetState() locks
		i, err := h.cluster.SyncApply(service.ApplyLogService.PrepareBackendDeclarationCommand(declaration))
		logger.Debug("cluster.SyncApply:", i, err)
	}

}

func (h *socketEventHandler) handleRoutesUpdate(conn ws.Conn, data []byte) {

}

func (h *socketEventHandler) handleModuleRequirements(conn ws.Conn, data []byte) {
	logger.Debugf("onReceivedModuleRequirements: %s %s", conn.Id(), string(data))

	instanceUuid, moduleName, err := conn.Parameters()
	logger.Debug("instanceUuid:", instanceUuid, "moduleName:", moduleName) // REMOVE
	if err != nil {
		_ = conn.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		return
	}

	declaration := bootstrap.ModuleRequirements{}
	err = json.Unmarshal(data, &declaration)
	if err != nil {
		_ = conn.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		logger.Debugf("onReceivedModuleRequirements: %s, error parse json data: %s", conn.Id(), err.Error())
		return
	}

	if declaration.RequireRoutes {
		// TODO h.store.GetState() locks
		service.DiscoveryService.SubscribeRoutes(conn, *h.store.GetState())
	}
	// TODO h.store.GetState() locks
	service.DiscoveryService.Subscribe(conn, declaration.RequiredModules, *h.store.GetState())
}

func (h *socketEventHandler) handleConfigSchema(conn ws.Conn, data []byte) {

}

func (h *socketEventHandler) handleRequestConfig(conn ws.Conn, data []byte) {

}
