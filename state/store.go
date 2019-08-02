package state

import (
	"github.com/hashicorp/raft"
	jsoniter "github.com/json-iterator/go"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"io"
	"sync"
)

var (
	json = jsoniter.ConfigFastest
)

type Store struct {
	state state
	lock  sync.Mutex
}

func (f *Store) Apply(l *raft.Log) interface{} {
	f.lock.Lock()
	defer f.lock.Unlock()
	return nil
}

// Snapshot returns a snapshot of the key-value store.
func (f *Store) Snapshot() (raft.FSMSnapshot, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	copied := deepcopy.Copy(f.state).(state)

	return &fsmSnapshot{copied}, nil
}

// Restore stores the key-value store to a previous state.
func (f *Store) Restore(rc io.ReadCloser) error {
	s := state{}

	if err := json.NewDecoder(rc).Decode(&s); err != nil {
		return errors.WithMessage(err, "unmarshal state")
	}

	f.state = s

	return nil
}

type fsmSnapshot struct {
	state state
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
	return &Store{
		state: newState(),
	}
}
