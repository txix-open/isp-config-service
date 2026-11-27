package tests_test

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
	backendTestModuleName = "test-module"
	backendTestModuleId   = "test-module-id"
	testConnId            = "test-conn-id"
	testNodeId            = "test-node-id"

	backendLeaderConnId   = "conn-1"
	backendFollowerConnId = "conn-2"
	backendLeaderAddress  = "10.0.0.1:9000"
	followerAddress       = "10.0.0.2:9000"
)

type backendStoredLog struct {
	message any
	fields  map[string]string
}

type backendLoggerSpy struct {
	warns []backendStoredLog
}

func (l *backendLoggerSpy) Warn(ctx context.Context, message any, fields ...log.Field) {
	entry := backendStoredLog{
		message: message,
		fields:  make(map[string]string, len(fields)),
	}
	for _, field := range fields {
		entry.fields[field.Key] = field.String
	}
	l.warns = append(l.warns, entry)
}

func (l *backendLoggerSpy) Error(ctx context.Context, message any, fields ...log.Field) {}

func (l *backendLoggerSpy) Info(ctx context.Context, message any, fields ...log.Field) {}

func (l *backendLoggerSpy) Debug(ctx context.Context, message any, fields ...log.Field) {}

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

type eventRepoSpy struct {
	inserted []entity.Event
}

func (r *eventRepoSpy) Insert(ctx context.Context, event entity.Event) error {
	r.inserted = append(r.inserted, event)
	return nil
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

			logger := &backendLoggerSpy{}
			backendRepo := &backendRepoSpy{}
			service := backend.NewBackend(backendRepo, &eventRepoSpy{}, testNodeId, nil, logger)

			declaration := cluster.BackendDeclaration{
				ModuleName:       backendTestModuleName,
				Version:          "1.0.0",
				LibVersion:       "1.0.0",
				Address:          cluster.AddressConfiguration{IP: "10.0.0.1", Port: "9000"},
				ModuleLeaderMode: tt.announcedMode,
			}
			connected, err := service.Connect(t.Context(), testConnId, backendTestModuleId, declaration)

			require.NoError(err)
			require.Len(backendRepo.upserted, 1)
			require.Equal(tt.expectedMode, backendRepo.upserted[0].ModuleLeaderMode)
			require.Equal(tt.expectedMode, connected.ModuleLeaderMode)

			if !tt.expectedWarn {
				require.Empty(logger.warns)
				return
			}

			require.Len(logger.warns, 1)
			require.Equal(backendTestModuleName, logger.warns[0].fields["moduleName"])
			require.Equal(tt.announcedMode, logger.warns[0].fields["moduleLeaderMode"])
		})
	}
}

func TestConnectInsertsOnlyLifecycleEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	backendRepo := newBackendRepoSpy(backendElectedInstance(backendLeaderConnId, backendLeaderAddress))
	service := backend.NewBackend(backendRepo, events, testNodeId, nil, &backendLoggerSpy{})

	declaration := cluster.BackendDeclaration{
		ModuleName:       backendTestModuleName,
		Address:          cluster.AddressConfiguration{IP: "10.0.0.2", Port: "9000"},
		ModuleLeaderMode: cluster.ModuleLeaderModeRequestElection,
	}
	_, err := service.Connect(t.Context(), backendFollowerConnId, backendTestModuleId, declaration)

	require.NoError(err)
	require.Equal([]string{"upsert"}, backendRepo.calls,
		"the connect path elects nobody: the leader is derived where the address list is emitted")
	require.Equal([]string{"module_lifecycle_" + backendTestModuleId}, eventKeys(events),
		"one lifecycle event per connect, no dedicated leader event")
}

func TestDisconnectInsertsOnlyLifecycleEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	backendRepo := newBackendRepoSpy(
		backendElectedInstance(backendLeaderConnId, backendLeaderAddress),
		backendElectedInstance(backendFollowerConnId, followerAddress),
	)
	service := backend.NewBackend(backendRepo, events, testNodeId, nil, &backendLoggerSpy{})

	err := service.Disconnect(t.Context(), entity.Backend{
		WsConnectionId: backendLeaderConnId,
		ModuleId:       backendTestModuleId,
		ModuleName:     backendTestModuleName,
	})

	require.NoError(err)
	require.Equal([]string{"delete"}, backendRepo.calls,
		"the disconnect path elects nobody")
	require.Equal([]string{"module_lifecycle_" + backendTestModuleId}, eventKeys(events),
		"one lifecycle event per disconnect, no dedicated leader event")
}

func TestClearPhantomBackendsInsertsOneLifecycleEventPerModule(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	backendRepo := newBackendRepoSpy(
		backendElectedInstance(backendLeaderConnId, backendLeaderAddress),
		backendElectedInstance(backendFollowerConnId, followerAddress),
	)
	service := backend.NewBackend(backendRepo, events, testNodeId, &etp.Rooms{}, &backendLoggerSpy{})

	deleted, err := service.ClearPhantomBackends(t.Context())

	require.NoError(err)
	require.Equal(2, deleted)
	require.Equal([]string{"own_node", "delete"}, backendRepo.calls,
		"the cleanup deletes and announces, it elects nobody")
	require.Equal([]string{"module_lifecycle_" + backendTestModuleId}, eventKeys(events),
		"the instances of one module are announced once")
}

func TestDeleteOwnBackendsInsertsNoEvents(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	events := &eventRepoSpy{}
	service := backend.NewBackend(newBackendRepoSpy(), events, testNodeId, nil, &backendLoggerSpy{})

	service.DeleteOwnBackends(t.Context())

	require.Empty(events.inserted, "graceful shutdown inserts no events at all")
}

func backendElectedInstance(connId string, address string) entity.Backend {
	return entity.Backend{
		WsConnectionId:   connId,
		ModuleId:         backendTestModuleId,
		ModuleName:       backendTestModuleName,
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
