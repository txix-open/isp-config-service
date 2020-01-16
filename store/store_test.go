package store

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"isp-config-service/entity"
	state2 "isp-config-service/store/state"
	"testing"
	"time"
)

type mockSnapshotSink struct {
	bytes.Buffer
}

func (d mockSnapshotSink) Close() error {
	return nil
}

func (d mockSnapshotSink) ID() string {
	panic("implement me")
}

func (d mockSnapshotSink) Cancel() error {
	panic("implement me")
}

func TestStore_SnapshotRestore(t *testing.T) {
	state := state2.NewState()
	state.WritableModules().Create(entity.Module{
		Id:                 state2.GenerateId(),
		Name:               "name1",
		CreatedAt:          time.Now().Round(time.Millisecond),
		LastConnectedAt:    time.Now().Round(time.Millisecond),
		LastDisconnectedAt: time.Now().Round(time.Millisecond),
	})

	store := NewStateStore(state)
	snapshot, err := store.Snapshot()
	assert.NoError(t, err)
	sink := mockSnapshotSink{}
	err = snapshot.Persist(&sink)
	assert.NoError(t, err)

	newState := state2.NewState()
	store.state = newState
	err = store.Restore(&sink)
	assert.NoError(t, err)

	assert.Equal(t, state, store.state)
}
