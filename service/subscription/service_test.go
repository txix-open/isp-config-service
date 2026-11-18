package subscription_test

import (
	"context"
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/subscription"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/etp/v4"
	"github.com/txix-open/isp-kit/log"
)

const (
	testModuleId   = "test-module-id"
	testModuleName = "test-module"

	leaderAddress    = "10.0.0.1:9000"
	successorAddress = "10.0.0.2:9000"
)

type storedLog struct {
	message any
	fields  map[string]string
}

type loggerSpy struct {
	infos []storedLog
}

func (l *loggerSpy) Info(ctx context.Context, message any, fields ...log.Field) {
	entry := storedLog{
		message: message,
		fields:  make(map[string]string, len(fields)),
	}
	for _, field := range fields {
		entry.fields[field.Key] = field.String
	}
	l.infos = append(l.infos, entry)
}

func (l *loggerSpy) Error(ctx context.Context, message any, fields ...log.Field) {}

func (l *loggerSpy) Warn(ctx context.Context, message any, fields ...log.Field) {}

func (l *loggerSpy) Debug(ctx context.Context, message any, fields ...log.Field) {}

type moduleRepoSpy struct{}

func (r *moduleRepoSpy) GetByNames(ctx context.Context, names []string) ([]entity.Module, error) {
	return nil, nil
}

func (r *moduleRepoSpy) GetById(ctx context.Context, id string) (*entity.Module, error) {
	return &entity.Module{Id: testModuleId, Name: testModuleName}, nil
}

type backendRepoSpy struct{}

func (r *backendRepoSpy) GetForLeaderElection(ctx context.Context, moduleId string) ([]entity.Backend, error) {
	return nil, nil
}

func (r *backendRepoSpy) All(ctx context.Context) ([]entity.Backend, error) {
	return nil, nil
}

type configRepoSpy struct{}

func (r *configRepoSpy) GetActive(ctx context.Context, moduleId string) (*entity.Config, error) {
	// The leader recompute path never reads a configuration.
	return &entity.Config{ModuleId: moduleId}, nil
}

type variableServiceSpy struct{}

func (s *variableServiceSpy) RenderConfig(ctx context.Context, input []byte) ([]byte, error) {
	return input, nil
}

type emitterSpy struct{}

func (e *emitterSpy) Emit(ctx context.Context, conn *etp.Conn, event string, data []byte) {}

func TestNotifyModuleLeaderRecomputeLogsLeaderChange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		recompute entity.PayloadModuleLeaderRecompute
		expected  *storedLog
		name      string
	}{
		{
			name: "the leader moved to the successor",
			recompute: entity.PayloadModuleLeaderRecompute{
				ModuleId:       testModuleId,
				ModuleName:     testModuleName,
				Reason:         entity.ModuleLeaderReasonDisconnect,
				PreviousLeader: leaderAddress,
				Leader:         successorAddress,
			},
			expected: &storedLog{
				message: "module leader changed",
				fields: map[string]string{
					"moduleName":            testModuleName,
					"previousLeaderAddress": leaderAddress,
					"newLeaderAddress":      successorAddress,
					"reason":                entity.ModuleLeaderReasonDisconnect,
				},
			},
		},
		{
			name: "the last leader is gone",
			recompute: entity.PayloadModuleLeaderRecompute{
				ModuleId:       testModuleId,
				ModuleName:     testModuleName,
				Reason:         entity.ModuleLeaderReasonDisconnect,
				PreviousLeader: leaderAddress,
				Leader:         "",
			},
			expected: &storedLog{
				message: "module leader changed",
				fields: map[string]string{
					"moduleName":            testModuleName,
					"previousLeaderAddress": leaderAddress,
					"newLeaderAddress":      "",
					"reason":                entity.ModuleLeaderReasonDisconnect,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			logger := &loggerSpy{}
			service := newService(logger)

			service.NotifyModuleLeaderRecompute(t.Context(), tt.recompute)

			require.Len(logger.infos, 1)
			require.Equal(tt.expected.message, logger.infos[0].message)
			require.Equal(tt.expected.fields, logger.infos[0].fields)
		})
	}
}

// Nothing the module leader recompute describes as a leader change means no log: the queue carries an event
// for every connect and disconnect, and most of them leave the leader where it was.
func TestNotifyModuleLeaderRecomputeSkipsUnchangedLeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		recompute entity.PayloadModuleLeaderRecompute
		name      string
	}{
		{
			name: "the leader is the same instance",
			recompute: entity.PayloadModuleLeaderRecompute{
				ModuleId:       testModuleId,
				ModuleName:     testModuleName,
				Reason:         entity.ModuleLeaderReasonReady,
				PreviousLeader: leaderAddress,
				Leader:         leaderAddress,
			},
		},
		{
			name: "the module has no leader at all",
			recompute: entity.PayloadModuleLeaderRecompute{
				ModuleId:   testModuleId,
				ModuleName: testModuleName,
				Reason:     entity.ModuleLeaderReasonReady,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			logger := &loggerSpy{}
			service := newService(logger)

			service.NotifyModuleLeaderRecompute(t.Context(), tt.recompute)

			require.Empty(logger.infos, "an event assembled without a leader change must not log")
		})
	}
}

func newService(logger log.Logger) subscription.Service {
	return subscription.NewService(
		&moduleRepoSpy{},
		&backendRepoSpy{},
		&configRepoSpy{},
		&etp.Rooms{},
		&emitterSpy{},
		&variableServiceSpy{},
		logger,
	)
}
