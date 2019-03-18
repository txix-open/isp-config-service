package service

import (
	"github.com/googollee/go-engine.io"
	"sync"
)

type StoreOption func(ss *sessionsStore)

type sessionsStore struct {
	sessions map[string]engineio.Conn
	locker   sync.RWMutex

	onRemove func(id string)
	onSet    func(id string, conn engineio.Conn)
}

func (s *sessionsStore) Get(id string) engineio.Conn {
	s.locker.RLock()
	defer s.locker.RUnlock()

	ret, ok := s.sessions[id]
	if !ok {
		return nil
	}

	return ret
}

func (s *sessionsStore) Set(id string, conn engineio.Conn) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.sessions[id] = conn

	if s.onSet != nil {
		s.onSet(id, conn)
	}
}

func (s *sessionsStore) Remove(id string) {
	s.locker.Lock()
	defer s.locker.Unlock()

	delete(s.sessions, id)

	if s.onRemove != nil {
		s.onRemove(id)
	}
}

func NewSessionStore(opts ...StoreOption) engineio.Sessions {
	ss := &sessionsStore{sessions: make(map[string]engineio.Conn)}

	for _, o := range opts {
		o(ss)
	}

	return ss
}

func WithOnRemoveCallback(f func(id string)) StoreOption {
	return func(ss *sessionsStore) {
		ss.onRemove = f
	}
}

func WithOnSetCallback(f func(id string, conn engineio.Conn)) StoreOption {
	return func(ss *sessionsStore) {
		ss.onSet = f
	}
}
