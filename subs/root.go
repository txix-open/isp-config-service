package subs

import (
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"
	"isp-config-service/ws"
)

const (
	Ok            = "ok"
	followersRoom = "followers"
)

var (
	json = jsoniter.ConfigFastest
)

type socketEventHandler struct {
	socket  *ws.WebsocketServer
	cluster *cluster.ClusterClient
	store   *store.Store
}

func (h *socketEventHandler) SubscribeAll() {
	h.socket.
		OnConnect(h.handleConnect).
		OnDisconnect(h.handleDisconnect).
		OnError(h.handleError).
		OnWithAck(cluster.ApplyCommandEvent, h.applyCommandOnLeader).
		OnWithAck(utils.ModuleReady, h.handleModuleReady).
		OnWithAck(utils.ModuleSendRequirements, h.handleModuleRequirements)
}

func (h *socketEventHandler) handleConnect(conn ws.Conn) {
	if conn.IsConfigClusterNode() {
		holder.Socket.Rooms().Join(conn, followersRoom)
	}
	moduleName, err := conn.Parameters()
	if err != nil {
		err := conn.Emit(utils.ErrorConnection, err.Error())
		if err != nil {
			log.Warnf(codes.SocketIoEmitError, "emit err %v", err)
		}
		return
	}
	command := cluster.PrepareModuleConnectedCommand(moduleName)

	applyLogResponse, err := h.cluster.SyncApply(command)
	if err != nil {
		log.Warnf(codes.SyncApplyError, "apply ModuleConnectedCommand: %v", err)
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		log.WithMetadata(map[string]interface{}{
			"comment":    applyLogResponse.Comment,
			"applyError": applyLogResponse.ApplyError,
		}).Warn(codes.SyncApplyError, "apply ModuleConnectedCommand")
	}

	var config map[string]interface{}
	h.store.VisitReadState(func(state state.ReadState) {
		config, err = state.GetCompiledConfig(moduleName)
	})
	if err != nil {
		err := conn.Emit(utils.ConfigError, err.Error())
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"module": moduleName,
			}).Warnf(codes.SocketIoEmitError, "emit err %v", err)
		}
		return
	}
	data, err := json.Marshal(config)
	if err != nil {
		err := conn.Emit(utils.ConfigError, err.Error())
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"module": moduleName,
			}).Warnf(codes.SocketIoEmitError, "emit err %v", err)
		}
		return
	}
	err = conn.Emit(utils.ConfigSendConfigWhenConnected, string(data))
	if err != nil {
		log.WithMetadata(map[string]interface{}{
			"module": moduleName,
		}).Warnf(codes.SocketIoEmitError, "emit config err %v", err)
	}

}

func (h *socketEventHandler) handleDisconnect(conn ws.Conn) {
	if conn.IsConfigClusterNode() {
		holder.Socket.Rooms().Leave(conn, followersRoom)
	}
	service.DiscoveryService.HandleDisconnect(conn.Id())
	service.RoutesService.HandleDisconnect(conn.Id())
	backend := conn.GetBackendDeclaration()
	if backend != nil {
		command := cluster.PrepareDeleteBackendDeclarationCommand(*backend)

		applyLogResponse, err := h.cluster.SyncApply(command)
		if err != nil {
			log.Warnf(codes.SyncApplyError, "apply ModuleConnectedCommand: %v", err)
		}
		if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
			log.WithMetadata(map[string]interface{}{
				"comment":    applyLogResponse.Comment,
				"applyError": applyLogResponse.ApplyError,
			}).Warn(codes.SyncApplyError, "apply DeleteBackendDeclarationCommand")
		}
	}
}

func (h *socketEventHandler) handleError(conn ws.Conn, err error) {
	log.Warnf(codes.SocketIoError, "socket.io: %v", err)
}

func NewSocketEventHandler(socket *ws.WebsocketServer, cluster *cluster.ClusterClient, store *store.Store) *socketEventHandler {
	return &socketEventHandler{
		socket:  socket,
		cluster: cluster,
		store:   store,
	}
}
