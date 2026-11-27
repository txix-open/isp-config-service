package tests_test

import (
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/event"

	"github.com/stretchr/testify/require"
)

const eventTestModuleId = "test-module-id"

func TestCompactCollapsesModuleLifecycle(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	ready := entity.NewEvent(entity.EventPayload{
		ModuleReady: &entity.PayloadModuleReady{ModuleId: eventTestModuleId},
	})
	ready.Id = 1

	disconnected := entity.NewEvent(entity.EventPayload{
		ModuleDisconnected: &entity.PayloadModuleDisconnected{ModuleId: eventTestModuleId},
	})
	disconnected.Id = 2

	configUpdated := entity.NewEvent(entity.EventPayload{
		ConfigUpdated: &entity.PayloadConfigUpdated{ModuleId: eventTestModuleId},
	})
	configUpdated.Id = 3

	require.Equal("module_lifecycle_"+eventTestModuleId, ready.Key())
	require.Equal("module_lifecycle_"+eventTestModuleId, disconnected.Key())
	require.Equal("config_updated_"+eventTestModuleId, configUpdated.Key())

	compacted := event.NewCompactor().Compact([]entity.Event{ready, disconnected, configUpdated})

	require.Equal([]entity.Event{disconnected, configUpdated}, compacted,
		"the newest lifecycle event wins, a config update is a key of its own")
}
