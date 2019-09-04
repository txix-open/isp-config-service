package subs

import (
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
	"isp-config-service/cluster"
	"isp-config-service/holder"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/ws"
)

const (
	ok = "ok"

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
		OnWithAck(cluster.ApplyCommandEvent, h.applyCommandOnLeader).
		On(utils.ModuleReady, h.handleModuleReady). //TODO сделать функцию с Ack, возвращать текстовку ошибки в случае возниконовения ошибки, константу ok, если все ок
		On(utils.ModuleSendRequirements, h.handleModuleRequirements)
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
	//TODO при отключении топология кластера меняется, соответсвенно нужно обновлять состояние и уведомлять всех об изменении
}

func (h *socketEventHandler) handleError(conn ws.Conn, err error) {
	logger.Warnf("socket.io: %v", err)
}

func NewSocketEventHandler(socket *ws.WebsocketServer, cluster *cluster.ClusterClient, store *store.Store) *socketEventHandler {
	return &socketEventHandler{
		socket:  socket,
		cluster: cluster,
		store:   store,
	}
}
