package service

import (
	"github.com/googollee/go-socket.io"
	"sync"
	"time"
)

type SocketConn struct {
	socketio.Socket
	EstablishedAt time.Time
}

type roomStats struct {
	m map[string]map[string]*SocketConn
	sync.RWMutex
}

func NewRoomStats() *roomStats {
	return &roomStats{
		m: make(map[string]map[string]*SocketConn),
	}
}

func (rs *roomStats) Join(room string, socket socketio.Socket) error {
	rs.Lock()
	sockets, ok := rs.m[room]
	if !ok {
		sockets = make(map[string]*SocketConn)
	}
	sockets[socket.Id()] = &SocketConn{socket, time.Now()}
	rs.m[room] = sockets
	rs.Unlock()
	return nil
}

func (rs *roomStats) GetConnection(instanceId, sockId string) (*SocketConn, bool) {
	rs.RLock()
	defer rs.RUnlock()

	if m, ok := rs.m[":"+instanceId]; !ok {
		return nil, false
	} else if s, ok := m[sockId]; !ok {
		return nil, false
	} else {
		return s, true
	}
}

func (rs *roomStats) Leave(room string, socket socketio.Socket) error {
	rs.Lock()
	defer rs.Unlock()
	sockets, ok := rs.m[room]
	if !ok {
		return nil
	}
	delete(sockets, socket.Id())
	if len(sockets) == 0 {
		delete(rs.m, room)
		return nil
	}
	rs.m[room] = sockets
	return nil
}

func (rs *roomStats) RoomCount(room string) map[string]*SocketConn {
	rs.RLock()
	defer rs.RUnlock()
	return rs.m[room]
}

func (rs *roomStats) RoomsCount() map[string]map[string]int {
	rs.RLock()
	defer rs.RUnlock()

	result := map[string]map[string]int{}
	for room, value := range rs.m {
		instanceId := room[1:37]
		module := room[37:]
		if module == "" {
			continue
		}
		if _, ok := result[instanceId]; !ok {
			result[instanceId] = map[string]int{module: len(value)}
		} else {
			result[instanceId][module] = len(value)
		}
	}
	return result
}

func (rs *roomStats) GetModuleConnectionMap() map[string]map[string][]string {
	rs.RLock()
	defer rs.RUnlock()

	result := make(map[string]map[string][]string)
	for room, value := range rs.m {
		instanceId := room[1:37]
		module := room[37:]
		if module == "" {
			continue
		}
		if _, ok := result[instanceId]; !ok {
			result[instanceId] = make(map[string][]string)
		}
		for sockId := range value {
			if arr, ok := result[instanceId][module]; !ok {
				result[instanceId][module] = []string{sockId}
			} else {
				result[instanceId][module] = append(arr, sockId)
			}
		}
	}
	return result
}

func (rs *roomStats) Send(ignore socketio.Socket, room, event string, args ...interface{}) error {
	rs.RLock()
	sockets := rs.m[room]
	for id, s := range sockets {
		if ignore != nil && ignore.Id() == id {
			continue
		}
		s.Emit(event, args...)
	}
	rs.RUnlock()
	return nil
}

func (rs *roomStats) Len(room string) int {
	return len(rs.m[room])
}

func (rs *roomStats) RemoveSocketConn(sockId string) {
	rs.Lock()

	for room, store := range rs.m {
		delete(store, sockId)
		if len(store) == 0 {
			delete(rs.m, room)
		}
	}

	rs.Unlock()
}
