package backend_test

import (
	"context"
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/module/backend"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
)

const (
	testModuleName = "test-module"
	testModuleId   = "test-module-id"
	testConnId     = "test-conn-id"
	testNodeId     = "test-node-id"
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
}

func (r *backendRepoSpy) Upsert(ctx context.Context, backend entity.Backend) error {
	r.upserted = append(r.upserted, backend)
	return nil
}

func (r *backendRepoSpy) DeleteByWsConnectionIds(ctx context.Context, wsConnIds []string) error {
	return nil
}

func (r *backendRepoSpy) GetByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) ([]entity.Backend, error) {
	return nil, nil
}

func (r *backendRepoSpy) DeleteByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) (int, error) {
	return 0, nil
}

type eventRepoSpy struct{}

func (r *eventRepoSpy) Insert(ctx context.Context, event entity.Event) error {
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
