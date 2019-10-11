package subs

import (
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/ws"
)

const (
	Ok            = "ok"
	followersRoom = "followers"
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
		OnWithAck(utils.ModuleSendConfigSchema, h.handleConfigSchema).
		OnWithAck(cluster.ApplyCommandEvent, h.applyCommandOnLeader).
		OnWithAck(utils.ModuleReady, h.handleModuleReady).
		OnWithAck(utils.ModuleSendRequirements, h.handleModuleRequirements)
}

func (h *socketEventHandler) handleConnect(conn ws.Conn) {
	if conn.IsConfigClusterNode() {
		holder.Socket.Rooms().Join(conn, followersRoom)
	} else {

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
		i, err := h.cluster.SyncApply(command)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"answer": i,
			}).Warnf(codes.SyncApplyError, "apply DeleteBackendDeclarationCommand %v", err)
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
