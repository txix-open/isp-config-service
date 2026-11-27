package tests_test

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
	subscriptionTestModuleId   = "test-module-id"
	subscriptionTestModuleName = "test-module"

	subscriptionLeaderConnId   = "conn-1"
	subscriptionFollowerConnId = "conn-2"

	subscriptionLeaderAddress = "10.0.0.1:9000"
	successorAddress          = "10.0.0.2:9000"
	plainAddress              = "10.0.0.3:9000"
)

type subscriptionStoredLog struct {
	message any
	fields  map[string]string
}

type subscriptionLoggerSpy struct {
	infos []subscriptionStoredLog
}

func (l *subscriptionLoggerSpy) Info(ctx context.Context, message any, fields ...log.Field) {
	entry := subscriptionStoredLog{
		message: message,
		fields:  make(map[string]string, len(fields)),
	}
	for _, field := range fields {
		entry.fields[field.Key] = field.String
	}
	l.infos = append(l.infos, entry)
}

func (l *subscriptionLoggerSpy) Error(ctx context.Context, message any, fields ...log.Field) {}

func (l *subscriptionLoggerSpy) Warn(ctx context.Context, message any, fields ...log.Field) {}

func (l *subscriptionLoggerSpy) Debug(ctx context.Context, message any, fields ...log.Field) {}

type moduleRepoSpy struct{}

func (r *moduleRepoSpy) GetByNames(ctx context.Context, names []string) ([]entity.Module, error) {
	return []entity.Module{{Id: subscriptionTestModuleId, Name: subscriptionTestModuleName}}, nil
}

func (r *moduleRepoSpy) GetById(ctx context.Context, id string) (*entity.Module, error) {
	return &entity.Module{Id: subscriptionTestModuleId, Name: subscriptionTestModuleName}, nil
}

type subscriptionBackendRepoSpy struct {
	forLeaderElection []entity.Backend
	all               []entity.Backend
}

func (r *subscriptionBackendRepoSpy) GetForLeaderElection(ctx context.Context, moduleId string) ([]entity.Backend, error) {
	return r.forLeaderElection, nil
}

func (r *subscriptionBackendRepoSpy) All(ctx context.Context) ([]entity.Backend, error) {
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

	logger := &subscriptionLoggerSpy{}
	rooms := etp.NewServer().Rooms()
	rooms.Join(&etp.Conn{}, subscription.BackendsChangingRoom(subscriptionTestModuleName))
	service := newService(logger, rooms, &subscriptionBackendRepoSpy{
		forLeaderElection: []entity.Backend{
			subscriptionElectedInstance(subscriptionLeaderConnId, subscriptionLeaderAddress),
			subscriptionElectedInstance(subscriptionFollowerConnId, successorAddress),
		},
	})

	err := service.NotifyBackendsChanged(t.Context(), subscriptionTestModuleId)

	require.NoError(err)
	require.Len(logger.infos, 1)
	require.Equal("module address list narrowed to the leader", logger.infos[0].message)
	require.Equal(subscriptionTestModuleName, logger.infos[0].fields["moduleName"])
	require.Equal(subscriptionLeaderAddress, logger.infos[0].fields["leaderAddress"],
		"the logged leader is the oldest elected instance, the one the subscriber received")
}

func TestNotifyRoutingChangedLogsPublishedLeaders(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := &subscriptionLoggerSpy{}
	rooms := etp.NewServer().Rooms()
	rooms.Join(&etp.Conn{}, subscription.RoutingChangingRoom())
	service := newService(logger, rooms, &subscriptionBackendRepoSpy{
		all: []entity.Backend{
			subscriptionElectedInstance(subscriptionLeaderConnId, subscriptionLeaderAddress),
			subscriptionElectedInstance(subscriptionFollowerConnId, successorAddress),
			plainInstance("conn-3", plainAddress),
		},
	})

	err := service.NotifyRoutingChanged(t.Context())

	require.NoError(err)
	require.Len(logger.infos, 1, "one summary line per routing rebuild")
	require.Equal("routing map publishes module leaders", logger.infos[0].message)
	require.Equal(subscriptionTestModuleName+"="+subscriptionLeaderAddress, logger.infos[0].fields["moduleLeaders"],
		"only the elected module and only its leader are published, the plain module is not listed")
}

func TestNotifyRoutingChangedWithoutElectionLogsNothing(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	logger := &subscriptionLoggerSpy{}
	rooms := etp.NewServer().Rooms()
	rooms.Join(&etp.Conn{}, subscription.RoutingChangingRoom())
	service := newService(logger, rooms, &subscriptionBackendRepoSpy{
		all: []entity.Backend{plainInstance("conn-3", plainAddress)},
	})

	err := service.NotifyRoutingChanged(t.Context())

	require.NoError(err)
	require.Empty(logger.infos, "modules without election keep the previous behaviour, including the log")
}

func subscriptionElectedInstance(connId string, address string) entity.Backend {
	return entity.Backend{
		WsConnectionId:   connId,
		ModuleId:         subscriptionTestModuleId,
		ModuleName:       subscriptionTestModuleName,
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
