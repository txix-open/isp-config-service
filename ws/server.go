package ws

import (
	"github.com/googollee/go-socket.io"
	"net/http"
)

type WebsocketServer struct {
	s         *socketio.Server
	roomStore *RoomStore
}

func (s *WebsocketServer) Close() error {
	return s.s.Close()
}

func (s *WebsocketServer) ServeHttp(w http.ResponseWriter, r *http.Request) {
	s.s.ServeHTTP(w, r)
}

func (s *WebsocketServer) Rooms() *RoomStore {
	return s.roomStore
}

func (s *WebsocketServer) Broadcast(except socketio.Conn, room string, msg string, v ...interface{}) error {
	conns := s.roomStore.ToBroadcast(except, room)
	for _, conn := range conns {
		conn.Emit(msg, v...)
	}
	return nil
}

func NewWebsocketServer() (*WebsocketServer, error) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	go server.Serve()

	return &WebsocketServer{
		s:         server,
		roomStore: NewRoomStore(),
	}, nil
}
