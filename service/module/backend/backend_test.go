package backend_test

import (
	"context"
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/module/backend"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/etp/v4"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
)

const (
	testModuleName = "test-module"
	testModuleId   = "test-module-id"
	testConnId     = "test-conn-id"
	testNodeId     = "test-node-id"

	leaderConnId    = "conn-1"
	followerConnId  = "conn-2"
	leaderAddress   = "10.0.0.1:9000"
	followerAddress = "10.0.0.2:9000"
)

type storedLog struct {
	message any
	fields  map[string]string
}

type loggerSpy struct {
	warns []storedLog
}

func (l *loggerSpy) Warn(ctx context.Context, message any, fields ...log.Field) {
	entry := storedLog{
		message: message,
		fields:  make(map[string]string, len(fields)),
	}
	for _, field := range fields {
		entry.fields[field.Key] = field.String
	}
	l.warns = append(l.warns, entry)
}

func (l *loggerSpy) Error(ctx context.Context, message any, fields ...log.Field) {}

func (l *loggerSpy) Info(ctx context.Context, message any, fields ...log.Field) {}

func (l *loggerSpy) Debug(ctx context.Context, message any, fields ...log.Field) {}

type backendRepoSpy struct {
	upserted []entity.Backend
	deleted  []string
	calls    []string
	rows     map[string]entity.Backend
}

func newBackendRepoSpy(backends ...entity.Backend) *backendRepoSpy {
	rows := make(map[string]entity.Backend, len(backends))
	for _, backend := range backends {
		rows[backend.WsConnectionId] = backend
	}

	return &backendRepoSpy{rows: rows}
}

func (r *backendRepoSpy) Upsert(ctx context.Context, backend entity.Backend) error {
	r.calls = append(r.calls, "upsert")
	r.upserted = append(r.upserted, backend)
	if r.rows == nil {
		r.rows = make(map[string]entity.Backend, 1)
	}
	r.rows[backend.WsConnectionId] = backend
	return nil
}

func (r *backendRepoSpy) DeleteByWsConnectionIds(ctx context.Context, wsConnIds []string) error {
	r.calls = append(r.calls, "delete")
	r.deleted = append(r.deleted, wsConnIds...)
	for _, wsConnId := range wsConnIds {
		delete(r.rows, wsConnId)
	}

	return nil
}

func (r *backendRepoSpy) GetByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) ([]entity.Backend, error) {
	r.calls = append(r.calls, "own_node")

	ownNode := make([]entity.Backend, 0, len(r.rows))
	for _, backend := range r.rows {
		ownNode = append(ownNode, backend)
	}

	return ownNode, nil
}

func (r *backendRepoSpy) DeleteByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) (int, error) {
	return 0, nil
}

func (r *backendRepoSpy) GetForLeaderElection(ctx context.Context, moduleId string) ([]entity.Backend, error) {
	r.calls = append(r.calls, "leader_election")

	moduleBackends := make([]entity.Backend, 0, len(r.rows))
	for _, backend := range r.rows {
		if backend.ModuleId == moduleId {
			moduleBackends = append(moduleBackends, backend)
		}
	}

	return moduleBackends, nil
}

type eventRepoSpy struct {
	inserted []entity.Event
}

func (r *eventRepoSpy) Insert(ctx context.Context, event entity.Event) error {
	r.inserted = append(r.inserted, event)
	return nil
}

func (r *eventRepoSpy) byKey(key string) []entity.Event {
	found := make([]entity.Event, 0, len(r.inserted))
	for _, event := range r.inserted {
		if event.Key() == key {
			found = append(found, event)
		}
	}

	return found
}

func TestConnectNormalizesModuleLeaderMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedMode  string
		announcedMode string
		name          string
		expectedWarn  bool
	}{
		{
			name:          "request election is stored as is",
			announcedMode: cluster.ModuleLeaderModeRequestElection,
			expectedMode:  cluster.ModuleLeaderModeRequestElection,
		},
		{
			name:          "empty mode is stored as empty",
			announcedMode: cluster.ModuleLeaderModeNone,
			expectedMode:  cluster.ModuleLeaderModeNone,
		},
		{
			name:          "im leader is reserved",
			announcedMode: cluster.ModuleLeaderModeImLeader,
			expectedMode:  cluster.ModuleLeaderModeNone,
			expectedWarn:  true,
		},
		{
			name:          "im follower is reserved",
			announcedMode: cluster.ModuleLeaderModeImFollower,
			expectedMode:  cluster.ModuleLeaderModeNone,
			expectedWarn:  true,
		},
		{
			name:          "garbage is unknown",
			announcedMode: "SOMETHING_ELSE",
			expectedMode:  cluster.ModuleLeaderModeNone,
			expectedWarn:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			logger := &loggerSpy{}
			backendRepo := &backendRepoSpy{}
			service := backend.NewBackend(backendRepo, &eventRepoSpy{}, testNodeId, nil, logger)

			declaration := cluster.BackendDeclaration{
				ModuleName:       testModuleName,
				Version:          "1.0.0",
				LibVersion:       "1.0.0",
				Address:          cluster.AddressConfiguration{IP: "10.0.0.1", Port: "9000"},
				ModuleLeaderMode: tt.announcedMode,
			}
			connected, err := service.Connect(t.Context(), testConnId, testModuleId, declaration)

			require.NoError(err)
			require.Len(backendRepo.upserted, 1)
			require.Equal(tt.expectedMode, backendRepo.upserted[0].ModuleLeaderMode)
			require.Equal(tt.expectedMode, connected.ModuleLeaderMode)

			if !tt.expectedWarn {
				require.Empty(logger.warns)
				return
			}

			require.Len(logger.warns, 1)
			require.Equal(testModuleName, logger.warns[0].fields["moduleName"])
			require.Equal(tt.announcedMode, logger.warns[0].fields["moduleLeaderMode"])
		})
	}
}

func TestConnectInsertsModuleLeaderRecomputeEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	backendRepo := newBackendRepoSpy(electedInstance(leaderConnId, leaderAddress))
	service := backend.NewBackend(backendRepo, events, testNodeId, nil, &loggerSpy{})

	declaration := cluster.BackendDeclaration{
		ModuleName: testModuleName,
		Address:    cluster.AddressConfiguration{IP: "10.0.0.2", Port: "9000"},
	}
	_, err := service.Connect(t.Context(), followerConnId, testModuleId, declaration)

	require.NoError(err)
	require.Equal([]string{"leader_election", "upsert", "leader_election"}, backendRepo.calls,
		"the leader must be derived from the stored instances before and after the connect")
	require.Equal([]string{
		"module_lifecycle_" + testModuleId,
		"module_leader_" + testModuleId,
	}, eventKeys(events), "the ready event and the leader recompute must both stay in the queue")

	recompute := requireRecompute(t, events, testModuleId)
	require.Equal(testModuleName, recompute.ModuleName)
	require.Equal(entity.ModuleLeaderReasonReady, recompute.Reason)
	require.Equal(leaderAddress, recompute.PreviousLeader)
	require.Equal(leaderAddress, recompute.Leader, "the oldest instance keeps leadership")
}

func TestConnectFirstInstanceInsertsModuleLeaderRecomputeEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	service := backend.NewBackend(newBackendRepoSpy(), events, testNodeId, nil, &loggerSpy{})

	declaration := cluster.BackendDeclaration{
		ModuleName:       testModuleName,
		Address:          cluster.AddressConfiguration{IP: "10.0.0.2", Port: "9000"},
		ModuleLeaderMode: cluster.ModuleLeaderModeRequestElection,
	}
	_, err := service.Connect(t.Context(), followerConnId, testModuleId, declaration)

	require.NoError(err)
	recompute := requireRecompute(t, events, testModuleId)
	require.Equal(entity.ModuleLeaderReasonReady, recompute.Reason)
	require.Empty(recompute.PreviousLeader, "the first instance of a module has no previous leader")
	require.Equal(followerAddress, recompute.Leader, "the first instance is the leader")
}

func TestDisconnectInsertsModuleLeaderRecomputeEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	backendRepo := newBackendRepoSpy(
		electedInstance(leaderConnId, leaderAddress),
		electedInstance(followerConnId, followerAddress),
	)
	service := backend.NewBackend(backendRepo, events, testNodeId, nil, &loggerSpy{})

	err := service.Disconnect(t.Context(), entity.Backend{
		WsConnectionId: leaderConnId,
		ModuleId:       testModuleId,
		ModuleName:     testModuleName,
	})

	require.NoError(err)
	require.Equal([]string{"leader_election", "delete", "leader_election"}, backendRepo.calls,
		"the leader must be derived from the stored instances before and after the disconnect")
	require.Equal([]string{
		"module_lifecycle_" + testModuleId,
		"module_leader_" + testModuleId,
	}, eventKeys(events), "the disconnected event and the leader recompute must both stay in the queue")

	recompute := requireRecompute(t, events, testModuleId)
	require.Equal(entity.ModuleLeaderReasonDisconnect, recompute.Reason)
	require.Equal(leaderAddress, recompute.PreviousLeader)
	require.Equal(followerAddress, recompute.Leader, "leadership moved to the surviving instance")
}

func TestClearPhantomBackendsInsertsModuleLeaderRecomputeEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	backendRepo := newBackendRepoSpy(electedInstance(leaderConnId, leaderAddress))
	service := backend.NewBackend(backendRepo, events, testNodeId, &etp.Rooms{}, &loggerSpy{})

	deleted, err := service.ClearPhantomBackends(t.Context())

	require.NoError(err)
	require.Equal(1, deleted)
	require.Equal([]string{"own_node", "leader_election", "delete", "leader_election"}, backendRepo.calls,
		"the leader must be derived from the stored instances before and after the phantom cleanup")
	require.Equal([]string{
		"module_lifecycle_" + testModuleId,
		"module_leader_" + testModuleId,
	}, eventKeys(events), "the disconnected event and the leader recompute must both stay in the queue")

	recompute := requireRecompute(t, events, testModuleId)
	require.Equal(testModuleId, recompute.ModuleId)
	require.Equal(testModuleName, recompute.ModuleName)
	require.Equal(entity.ModuleLeaderReasonDisconnect, recompute.Reason)
	require.Equal(leaderAddress, recompute.PreviousLeader)
	require.Empty(recompute.Leader, "with no instances left the module has no leader")
}

func TestDeleteOwnBackendsInsertsNoEvents(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	service := backend.NewBackend(newBackendRepoSpy(), events, testNodeId, nil, &loggerSpy{})

	service.DeleteOwnBackends(t.Context())

	require.Empty(events.inserted, "graceful shutdown keeps its current behaviour: no leader recompute")
}

func electedInstance(connId string, address string) entity.Backend {
	return entity.Backend{
		WsConnectionId:   connId,
		ModuleId:         testModuleId,
		ModuleName:       testModuleName,
		Address:          address,
		ModuleLeaderMode: cluster.ModuleLeaderModeRequestElection,
	}
}

func eventKeys(events *eventRepoSpy) []string {
	keys := make([]string, 0, len(events.inserted))
	for _, event := range events.inserted {
		keys = append(keys, event.Key())
	}

	return keys
}

// requireRecompute returns the single leader recompute the module produced, proving it has a compaction key
// of its own next to the lifecycle event of the same module. Every case in this file uses one module, so the
// module id is a parameter only to keep the assertions readable at the call site.
// nolint:unparam
func requireRecompute(
	t *testing.T,
	events *eventRepoSpy,
	moduleId string,
) *entity.PayloadModuleLeaderRecompute {
	t.Helper()
	require := require.New(t)

	lifecycleKey := "module_lifecycle_" + testModuleId
	recomputeKey := "module_leader_" + testModuleId
	require.NotEqual(lifecycleKey, recomputeKey)

	found := events.byKey(recomputeKey)
	require.Len(found, 1, "the leader recompute must have its own compaction key")

	recompute := found[0].Payload.Value.ModuleLeaderRecompute
	require.NotNil(recompute)
	require.Equal(testModuleId, recompute.ModuleId)

	return recompute
}
