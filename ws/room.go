package ws

import (
	socketio "github.com/googollee/go-socket.io"
	"sync"
)

type RoomStore struct {
	mu    sync.RWMutex
	rooms map[string]map[string]socketio.Conn
}

func (s *RoomStore) Join(conn socketio.Conn, rooms ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			conns[conn.ID()] = conn
		} else {
			s.rooms[room] = map[string]socketio.Conn{conn.ID(): conn}
		}
	}
}

func (s *RoomStore) Leave(conn socketio.Conn, rooms ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			delete(conns, conn.ID())
			if len(conns) == 0 {
				delete(s.rooms, room)
			}
		}
	}
}

func (s *RoomStore) ToBroadcast(except socketio.Conn, rooms ...string) []socketio.Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]socketio.Conn, 0)
	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			for id, conn := range conns {
				if id != except.ID() {
					result = append(result, conn)
				}
			}
		}
	}

	return result
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		mu:    sync.RWMutex{},
		rooms: make(map[string]map[string]socketio.Conn),
	}
}
