package ws

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/cluster"
)

type Conn interface {
	Id() string
	Emit(event string, args ...interface{}) error
	IsConfigClusterNode() bool
	SetBackendDeclaration(backend structure.BackendDeclaration)
	GetBackendDeclaration() *structure.BackendDeclaration
}

type wsConn struct {
	conn    socketio.Socket
	backend *structure.BackendDeclaration
}

func (c *wsConn) Id() string {
	return c.conn.Id()
}

func (c *wsConn) Emit(event string, args ...interface{}) error {
	return c.conn.Emit(event, args...)
}

func (c *wsConn) IsConfigClusterNode() bool {
	isCluster := c.conn.Request().URL.Query().Get(cluster.ClusterParam)
	return isCluster == "true"
}

func (c *wsConn) SetBackendDeclaration(backend structure.BackendDeclaration) {
	c.backend = &backend
}

func (c *wsConn) GetBackendDeclaration() *structure.BackendDeclaration {
	return c.backend
}
