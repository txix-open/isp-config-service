package subscription_test

import (
	"context"
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/subscription"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/etp/v4"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
)

const (
	testModuleId   = "test-module-id"
	testModuleName = "test-module"

	leaderConnId   = "conn-1"
	followerConnId = "conn-2"

	leaderAddress    = "10.0.0.1:9000"
	successorAddress = "10.0.0.2:9000"
	plainAddress     = "10.0.0.3:9000"
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
	return []entity.Module{{Id: testModuleId, Name: testModuleName}}, nil
}

func (r *moduleRepoSpy) GetById(ctx context.Context, id string) (*entity.Module, error) {
	return &entity.Module{Id: testModuleId, Name: testModuleName}, nil
}

type backendRepoSpy struct {
	forLeaderElection []entity.Backend
	all               []entity.Backend
}

func (r *backendRepoSpy) GetForLeaderElection(ctx context.Context, moduleId string) ([]entity.Backend, error) {
	return r.forLeaderElection, nil
}

func (r *backendRepoSpy) All(ctx context.Context) ([]entity.Backend, error) {
	return r.all, nil
}

type configRepoSpy struct{}

func (r *configRepoSpy) GetActive(ctx context.Context, moduleId string) (*entity.Config, error) {
	return &entity.Config{ModuleId: moduleId}, nil
}

type variableServiceSpy struct{}

func (s *variableServiceSpy) RenderConfig(ctx context.Context, moduleId string, input []byte) ([]byte, error) {
	return input, nil
}

type emitterSpy struct{}

func (e *emitterSpy) Emit(ctx context.Context, conn *etp.Conn, event string, data []byte) {}

// The on-call engineer answers "who is accepting sessions of this module" from the log of the node that
// emitted the address list: the same derivation the subscriber received, logged where it is published.
func TestNotifyBackendsChangedLogsTheLeaderItPublishes(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := &loggerSpy{}
	rooms := etp.NewServer().Rooms()
	rooms.Join(&etp.Conn{}, subscription.BackendsChangingRoom(testModuleName))
	service := newService(logger, rooms, &backendRepoSpy{
		forLeaderElection: []entity.Backend{
			electedInstance(leaderConnId, leaderAddress),
			electedInstance(followerConnId, successorAddress),
		},
	})

	err := service.NotifyBackendsChanged(t.Context(), testModuleId)

	require.NoError(err)
	require.Len(logger.infos, 1)
	require.Equal("module address list narrowed to the leader", logger.infos[0].message)
	require.Equal(testModuleName, logger.infos[0].fields["moduleName"])
	require.Equal(leaderAddress, logger.infos[0].fields["leaderAddress"],
		"the logged leader is the oldest elected instance, the one the subscriber received")
}

func TestNotifyRoutingChangedLogsPublishedLeaders(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := &loggerSpy{}
	rooms := etp.NewServer().Rooms()
	rooms.Join(&etp.Conn{}, subscription.RoutingChangingRoom())
	service := newService(logger, rooms, &backendRepoSpy{
		all: []entity.Backend{
			electedInstance(leaderConnId, leaderAddress),
			electedInstance(followerConnId, successorAddress),
			plainInstance("conn-3", plainAddress),
		},
	})

	err := service.NotifyRoutingChanged(t.Context())

	require.NoError(err)
	require.Len(logger.infos, 1, "one summary line per routing rebuild")
	require.Equal("routing map publishes module leaders", logger.infos[0].message)
	require.Equal(testModuleName+"="+leaderAddress, logger.infos[0].fields["moduleLeaders"],
		"only the elected module and only its leader are published, the plain module is not listed")
}

func TestNotifyRoutingChangedWithoutElectionLogsNothing(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := &loggerSpy{}
	rooms := etp.NewServer().Rooms()
	rooms.Join(&etp.Conn{}, subscription.RoutingChangingRoom())
	service := newService(logger, rooms, &backendRepoSpy{
		all: []entity.Backend{plainInstance("conn-3", plainAddress)},
	})

	err := service.NotifyRoutingChanged(t.Context())

	require.NoError(err)
	require.Empty(logger.infos, "modules without election keep the previous behaviour, including the log")
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

func plainInstance(connId string, address string) entity.Backend {
	return entity.Backend{
		WsConnectionId: connId,
		ModuleId:       "plain-module-id",
		ModuleName:     "plain-module",
		Address:        address,
	}
}

func newService(logger log.Logger, rooms *etp.Rooms, backendRepo subscription.BackendRepo) subscription.Service {
	return subscription.NewService(
		&moduleRepoSpy{},
		backendRepo,
		&configRepoSpy{},
		rooms,
		&emitterSpy{},
		&variableServiceSpy{},
		logger,
	)
}
