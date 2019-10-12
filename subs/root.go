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
		OnWithAck(utils.ModuleSendRequirements, h.handleModuleRequirements).
		OnWithAck(utils.ModuleSendConfigSchema, h.handleConfigSchema)
}

func (h *socketEventHandler) handleConnect(conn ws.Conn) {
	if conn.IsConfigClusterNode() {
		holder.Socket.Rooms().Join(conn, followersRoom)
	}
	moduleName, err := conn.Parameters()
	if err != nil {
		EmitConn(conn, utils.ErrorConnection, err.Error())
		return
	}
	command := cluster.PrepareModuleConnectedCommand(moduleName)
	h.SyncApplyCommand(command, "ModuleConnectedCommand")

	var config map[string]interface{}
	h.store.VisitReadState(func(state state.ReadState) {
		config, err = state.GetCompiledConfig(moduleName)
	})
	if err != nil {
		EmitConn(conn, utils.ConfigError, err.Error())
		return
	}
	data, err := json.Marshal(config)
	if err != nil {
		EmitConn(conn, utils.ConfigError, err.Error())
		return
	}
	EmitConn(conn, utils.ConfigSendConfigWhenConnected, string(data))
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
		h.SyncApplyCommand(command, "ModuleConnectedCommand")
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
