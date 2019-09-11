package ws

import (
	"sync"
)

const (
	idsRoom = "__id"
)

type RoomStore struct {
	mu    sync.RWMutex
	rooms map[string]map[string]Conn
}

func (s *RoomStore) GetOrJoinById(id string, new Conn) Conn {
	s.mu.RLock()
	var (
		conn Conn
		ok   bool
	)
	idRoom, roomExist := s.rooms[idsRoom]
	if roomExist {
		conn, ok = idRoom[id]
	}
	s.mu.RUnlock()
	if ok {
		return conn
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	idRoom, roomExist = s.rooms[idsRoom]
	if roomExist {
		conn, ok = idRoom[id]
	}
	if ok {
		return conn
	}
	conn = new
	if !roomExist {
		s.rooms[idsRoom] = map[string]Conn{conn.Id(): conn}
		return conn
	}
	idRoom[conn.Id()] = conn
	return conn
}

func (s *RoomStore) Join(conn Conn, rooms ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			conns[conn.Id()] = conn
		} else {
			s.rooms[room] = map[string]Conn{conn.Id(): conn}
		}
	}
}

func (s *RoomStore) Leave(conn Conn, rooms ...string) {
	s.LeaveByConnId(conn.Id(), rooms...)
}

func (s *RoomStore) LeaveByConnId(id string, rooms ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			delete(conns, id)
			if len(conns) == 0 {
				delete(s.rooms, room)
			}
		}
	}
}

func (s *RoomStore) ToBroadcast(rooms ...string) []Conn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Conn, 0)
	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			for _, conn := range conns {
				result = append(result, conn)
			}
		}
	}

	return result
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		mu:    sync.RWMutex{},
		rooms: make(map[string]map[string]Conn),
	}
}
