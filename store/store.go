package store

import (
	"encoding/binary"
	"github.com/hashicorp/raft"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"io"
	"isp-config-service/cluster"
	"isp-config-service/codes"
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
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(l.Data) < 8 {
		log.Fatalf(codes.ApplyLogCommandError, "invalid log data command: %s", l.Data)
	}
	command := binary.BigEndian.Uint64(l.Data[:8])
	if handler, ok := s.handlers[command]; ok {
		err := handler(l.Data[8:])
		if err != nil {
			log.Fatalf(codes.ApplyLogCommandError, "apply log command: %v", err)
		}
	} else {
		log.Fatalf(codes.ApplyLogCommandError, "unknown log command %s", command)
	}
	return nil
}

func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	s.lock.Lock()
	copied := deepcopy.Copy(s.state).(state.State)
	s.lock.Unlock()
	return &fsmSnapshot{copied}, nil
}

func (s *Store) Restore(rc io.ReadCloser) error {
	state2 := state.State{}
	if err := json.NewDecoder(rc).Decode(&state2); err != nil {
		return errors.WithMessage(err, "unmarshal store")
	}
	s.state = state2
	return nil
}

func (s *Store) GetReadState() state.ReadState {
	return &s.state
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

func NewStateStore(st state.State) *Store {
	store := &Store{
		state: st,
	}
	store.handlers = map[uint64]func([]byte) error{
		cluster.UpdateBackendDeclarationCommand: store.applyUpdateBackendDeclarationCommand,
		cluster.DeleteBackendDeclarationCommand: store.applyDeleteBackendDeclarationCommand,
		cluster.UpdateConfigSchemaCommand:       store.applyUpdateConfigSchema,
	}
	return store
}
