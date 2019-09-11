package ws

import (
	"github.com/googollee/go-socket.io"
	gosocketio "github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/logger"
	"net/http"
)

type WebsocketServer struct {
	s         *socketio.Server
	roomStore *RoomStore
}

func (s *WebsocketServer) Close() error {
	return nil
}

func (s *WebsocketServer) ServeHttp(w http.ResponseWriter, r *http.Request) {
	s.s.ServeHTTP(w, r)
}

func (s *WebsocketServer) Rooms() *RoomStore {
	return s.roomStore
}

func (s *WebsocketServer) OnWithAck(event string, f func(conn Conn, data []byte) string) *WebsocketServer {
	must(s.s.On(event, func(conn socketio.Socket, data []byte) string {
		logger.Debugf("[%s]:%s:%s", conn.Request().RemoteAddr, event, string(data))

		return f(s.roomStore.GetOrJoinById(conn.Id(), &wsConn{conn: conn}), data)
	}))
	return s
}

func (s *WebsocketServer) On(event string, f func(conn Conn, data []byte)) *WebsocketServer {
	must(s.s.On(event, func(conn socketio.Socket, data []byte) {
		logger.Debugf("[%s]:%s:%s", conn.Request().RemoteAddr, event, string(data))

		f(s.roomStore.GetOrJoinById(conn.Id(), &wsConn{conn: conn}), data)
	}))
	return s
}

func (s *WebsocketServer) OnConnect(f func(Conn)) *WebsocketServer {
	must(s.s.On(gosocketio.OnConnection, func(conn socketio.Socket) {
		logger.Debugf("[%s]:%s", conn.Request().RemoteAddr, gosocketio.OnConnection)

		f(s.roomStore.GetOrJoinById(conn.Id(), &wsConn{conn: conn}))
	}))
	return s
}

func (s *WebsocketServer) OnDisconnect(f func(Conn)) *WebsocketServer {
	must(s.s.On(gosocketio.OnDisconnection, func(conn socketio.Socket) {
		logger.Debugf("[%s]:%s", conn.Request().RemoteAddr, gosocketio.OnDisconnection)

		c := s.roomStore.GetOrJoinById(conn.Id(), &wsConn{conn: conn})
		f(c)
		s.roomStore.Leave(c, idsRoom)
	}))
	return s
}

func (s *WebsocketServer) OnError(f func(Conn, error)) *WebsocketServer {
	must(s.s.On(gosocketio.OnError, func(conn socketio.Socket, err error) {
		logger.Debugf("[%s]:%s:%v", conn.Request().RemoteAddr, gosocketio.OnError, err)

		f(s.roomStore.GetOrJoinById(conn.Id(), &wsConn{conn: conn}), err)
	}))
	return s
}

func (s *WebsocketServer) Broadcast(room string, msg string, v ...interface{}) error {
	conns := s.roomStore.ToBroadcast(room)
	for _, conn := range conns {
		err := conn.Emit(msg, v...)
		if err != nil {
			logger.Warnf("Broadcast to %s, message %s. err: %v", room, msg, err)
		}
	}
	return nil
}

func NewWebsocketServer() (*WebsocketServer, error) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	return &WebsocketServer{
		s:         server,
		roomStore: NewRoomStore(),
	}, nil
}
