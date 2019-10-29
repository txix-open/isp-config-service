package subs

import (
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/entity"
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
	socket *ws.WebsocketServer
	store  *store.Store
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
	log.Debugf(0, "handleConnect: %s", moduleName) // REMOVE
	if err != nil {
		EmitConn(conn, utils.ErrorConnection, err.Error())
		conn.Disconnect()
		return
	}
	holder.Socket.Rooms().Join(conn, moduleName+service.ConfigWatchersRoomSuffix)
	now := state.GenerateDate()
	module := entity.Module{
		Id:              state.GenerateId(),
		Name:            moduleName,
		CreatedAt:       now,
		LastConnectedAt: now,
	}
	command := cluster.PrepareModuleConnectedCommand(module)
	_, err = SyncApplyCommand(command, "ModuleConnectedCommand")
	if err != nil {
		EmitConn(conn, utils.ErrorConnection, err.Error())
		conn.Disconnect()
		return
	}

	var config map[string]interface{}
	h.store.VisitReadonlyState(func(state state.ReadonlyState) {
		config, err = service.ConfigService.GetCompiledConfig(moduleName, state)
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
	moduleName, _ := conn.Parameters()
	log.Debugf(0, "handleDisconnect: %s", moduleName) // REMOVE
	if moduleName != "" {
		holder.Socket.Rooms().Leave(conn, moduleName+service.ConfigWatchersRoomSuffix)
		now := state.GenerateDate()
		module := entity.Module{
			Id:                 state.GenerateId(),
			Name:               moduleName,
			CreatedAt:          now,
			LastConnectedAt:    now,
			LastDisconnectedAt: now,
		}
		command := cluster.PrepareModuleDisconnectedCommand(module)
		_, _ = SyncApplyCommand(command, "ModuleDisconnectedCommand")

	}
	service.DiscoveryService.HandleDisconnect(conn.Id())
	service.RoutesService.HandleDisconnect(conn.Id())
	backend := conn.GetBackendDeclaration()
	if backend != nil {
		command := cluster.PrepareDeleteBackendDeclarationCommand(*backend)
		_, _ = SyncApplyCommand(command, "DeleteBackendDeclarationCommand")
	}
}

func (h *socketEventHandler) handleError(conn ws.Conn, err error) {
	log.Warnf(codes.SocketIoError, "socket.io: %v", err)
}

func NewSocketEventHandler(socket *ws.WebsocketServer, store *store.Store) *socketEventHandler {
	return &socketEventHandler{
		socket: socket,
		store:  store,
	}
}
