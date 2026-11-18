package event_test

import (
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/event"

	"github.com/stretchr/testify/require"
)

const (
	testModuleId   = "test-module-id"
	testModuleName = "test-module"
)

func TestCompactKeepsModuleLeaderRecompute(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	ready := entity.NewEvent(entity.EventPayload{
		ModuleReady: &entity.PayloadModuleReady{ModuleId: testModuleId},
	})
	ready.Id = 1

	disconnected := entity.NewEvent(entity.EventPayload{
		ModuleDisconnected: &entity.PayloadModuleDisconnected{ModuleId: testModuleId},
	})
	disconnected.Id = 2

	recompute := moduleLeaderRecomputeEvent(entity.ModuleLeaderReasonDisconnect, "10.0.0.1:9000")
	recompute.Id = 3

	require.Equal("module_lifecycle_"+testModuleId, ready.Key())
	require.Equal("module_lifecycle_"+testModuleId, disconnected.Key())
	require.Equal("module_leader_"+testModuleId, recompute.Key(),
		"the leader recompute must not share the lifecycle key")
	require.NotEmpty(recompute.Key())

	compacted := event.NewCompactor().Compact([]entity.Event{ready, disconnected, recompute})

	require.Equal([]entity.Event{disconnected, recompute}, compacted,
		"the newest lifecycle event wins, the leader recompute is a key of its own")
}

func moduleLeaderRecomputeEvent(reason string, previousLeader string) entity.Event {
	return entity.NewEvent(entity.EventPayload{
		ModuleLeaderRecompute: &entity.PayloadModuleLeaderRecompute{
			ModuleId:       testModuleId,
			ModuleName:     testModuleName,
			Reason:         reason,
			PreviousLeader: previousLeader,
		},
	})
}
