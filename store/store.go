package store

import (
	"encoding/binary"
	"github.com/hashicorp/raft"
	"github.com/integration-system/isp-lib/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"io"
	"isp-config-service/cluster"
	"isp-config-service/store/state"
	"sync"
)

var (
	json = jsoniter.ConfigFastest
)

type Store struct {
	state    state.State
	lock     sync.RWMutex
	handlers map[uint64]func([]byte) error
}

func (s *Store) Apply(l *raft.Log) interface{} {
	logger.Info("Applying...")
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(l.Data) < 8 {
		logger.Fatal("Invalid log data command", l.Data)
	}
	command := binary.BigEndian.Uint64(l.Data[:8])
	if handler, ok := s.handlers[command]; ok {
		err := handler(l.Data[8:])
		if err != nil {
			logger.Fatal("error while applying log command", command, err)
		}
	} else {
		logger.Fatal("Log command not found", command)
	}
	return nil
}

func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	logger.Debug("Create snapshot")

	copied := deepcopy.Copy(s.state).(state.State)
	return &fsmSnapshot{copied}, nil
}

func (s *Store) Restore(rc io.ReadCloser) error {
	state := state.State{}
	logger.Debug("Try to restore snapshot")
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return errors.WithMessage(err, "unmarshal store")
	}
	s.state = state
	return nil
}

func (s *Store) GetReadState() state.ReadState {
	return s.state
}

func (s *Store) VisitReadState(f func(state.ReadState)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	f(s.GetReadState())
}

func (s *Store) VisitState(f func(state.State)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	f(s.state)
}

type fsmSnapshot struct {
	state state.State
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b, err := json.Marshal(f.state)
		if err != nil {
			return err
		}
		if _, err := sink.Write(b); err != nil {
			return err
		}
		return sink.Close()
	}()

	if err != nil {
		_ = sink.Cancel()
	}
	return err
}

func (f *fsmSnapshot) Release() {}

func NewStateStore() *Store {
	store := &Store{
		state: state.NewState(),
	}
	store.handlers = map[uint64]func([]byte) error{
		cluster.UpdateBackendDeclarationCommand: store.applyUpdateBackendDeclarationCommand,
		cluster.DeleteBackendDeclarationCommand: store.applyDeleteBackendDeclarationCommand,
	}
	return store
}
