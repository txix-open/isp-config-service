package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/store/state"
)

var (
	json = jsoniter.ConfigFastest
)

type Store struct {
	state    *state.State
	lock     sync.RWMutex
	handlers map[uint64]func([]byte) (interface{}, error)
}

func (s *Store) Apply(l *raft.Log) interface{} {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(l.Data) < cluster.CommandSizeBytes {
		log.Errorf(codes.ApplyLogCommandError, "invalid log data command: %s", l.Data)
	}
	command := binary.BigEndian.Uint64(l.Data[:8])
	log.Debugf(0, "Apply %d command. Data: %s", command, l.Data)

	var (
		result interface{}
		err    error
	)
	if handler, ok := s.handlers[command]; ok {
		result, err = handler(l.Data[8:])
	} else {
		err = fmt.Errorf("unknown log command %d", command)
		log.WithMetadata(map[string]interface{}{
			"command": command,
			"body":    string(l.Data),
		}).Error(codes.ApplyLogCommandError, "unknown log command")
	}

	bytes, e := json.Marshal(result)
	if e != nil {
		panic(e) // must never occurred
	}

	logResponse := cluster.ApplyLogResponse{Result: bytes}
	if err != nil {
		logResponse.ApplyError = err.Error()
	}
	return logResponse
}

func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	s.lock.Lock()
	copied := deepcopy.Copy(s.state).(*state.State)
	s.lock.Unlock()
	return &fsmSnapshot{copied}, nil
}

func (s *Store) Restore(rc io.ReadCloser) error {
	state2 := state.State{}
	if err := json.NewDecoder(rc).Decode(&state2); err != nil {
		return errors.WithMessage(err, "unmarshal store")
	}
	s.state = &state2
	return nil
}

func (s *Store) VisitReadonlyState(f func(state.ReadonlyState)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	f(*s.state)
}

func (s *Store) VisitState(f func(writableState state.WritableState)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	f(s.state)
}

type fsmSnapshot struct {
	state *state.State
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

func NewStateStore(st *state.State) *Store {
	store := &Store{
		state: st,
	}
	store.handlers = map[uint64]func([]byte) (interface{}, error){
		cluster.UpdateBackendDeclarationCommand: store.applyUpdateBackendDeclarationCommand,
		cluster.DeleteBackendDeclarationCommand: store.applyDeleteBackendDeclarationCommand,
		cluster.ModuleConnectedCommand:          store.applyModuleConnectedCommand,
		cluster.ModuleDisconnectedCommand:       store.applyModuleDisconnectedCommand,
		cluster.DeleteModulesCommand:            store.applyDeleteModulesCommand,
		cluster.UpdateConfigSchemaCommand:       store.applyUpdateConfigSchemaCommand,
		cluster.ActivateConfigCommand:           store.applyActivateConfigCommand,
		cluster.DeleteConfigsCommand:            store.applyDeleteConfigsCommand,
		cluster.UpsertConfigCommand:             store.applyUpsertConfigCommand,
		cluster.DeleteCommonConfigsCommand:      store.applyDeleteCommonConfigsCommand,
		cluster.UpsertCommonConfigCommand:       store.applyUpsertCommonConfigCommand,
	}
	return store
}
