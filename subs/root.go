package subs

import (
	"github.com/integration-system/isp-lib/logger"
	"isp-config-service/cluster"
	"isp-config-service/holder"
	"isp-config-service/state"
	"isp-config-service/ws"
)

const (
	ok = "ok"

	followersRoom = "followers"
)

type socketEventHandler struct {
	socket  *ws.WebsocketServer
	cluster *cluster.ClusterClient
	state   *state.Store
}

func (h *socketEventHandler) SubscribeAll() {
	h.socket.
		OnConnect(h.handleConnect).
		OnDisconnect(h.handleDisconnect).
		OnError(h.handleError).
		OnWithAck(cluster.ApplyCommandEvent, h.applyCommandOnLeader)
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
}

func (h *socketEventHandler) handleError(conn ws.Conn, err error) {
	logger.Warnf("socket.io: %v", err)
}

func NewSocketEventHandler(socket *ws.WebsocketServer, cluster *cluster.ClusterClient, state *state.Store) *socketEventHandler {
	return &socketEventHandler{
		socket:  socket,
		cluster: cluster,
		state:   state,
	}
}
