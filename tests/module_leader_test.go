package tests_test

import (
	"slices"
	"testing"
	"time"

	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/service/module/leader"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
)

const (
	electionMode   = cluster.ModuleLeaderModeRequestElection
	noElectionMode = cluster.ModuleLeaderModeNone
	garbageMode    = "SOMETHING_ELSE"
)

func TestDerive(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	elected := leader.Derive(instances(
		instance("conn-c", 100, noElectionMode),
		instance("conn-a", 200, electionMode),
		instance("conn-b", 100, noElectionMode),
	))

	require.True(elected.HasElection)
	require.True(elected.Found)
	require.Equal("conn-b", elected.Leader.WsConnectionId, "leader is the earliest instance, tie broken by ws_connection_id")

	first := leader.Derive(instances(
		instance("conn-c", 100, noElectionMode),
		instance("conn-a", 200, electionMode),
		instance("conn-b", 100, noElectionMode),
	))
	second := leader.Derive(instances(
		instance("conn-a", 200, electionMode),
		instance("conn-b", 100, noElectionMode),
		instance("conn-c", 100, noElectionMode),
	))

	require.Equal(first, second, "the same set of instances must yield the same leader")
	require.Equal(elected.Leader.WsConnectionId, first.Leader.WsConnectionId)
	require.Equal(elected.Leader.WsConnectionId, second.Leader.WsConnectionId)
}

func TestDerive_InputOrder(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	base := instances(
		instance("conn-d", 100, electionMode),
		instance("conn-a", 300, electionMode),
		instance("conn-c", 100, noElectionMode),
		instance("conn-b", 200, noElectionMode),
	)

	expected := leader.Derive(base)
	require.Equal("conn-c", expected.Leader.WsConnectionId)

	for _, input := range permutations(base) {
		original := slices.Clone(base)

		got := leader.Derive(input)

		require.Equal(expected, got, "input order must not change the result")
		require.Equal(original, base, "Derive must not mutate the caller slice")
	}
}

func TestDerive_NoElection(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	got := leader.Derive(instances(
		instance("conn-a", 100, noElectionMode),
		instance("conn-b", 200, cluster.ModuleLeaderModeImLeader),
		instance("conn-c", 50, garbageMode),
	))

	require.False(got.HasElection)
	require.False(got.Found)
	require.Equal(entity.Backend{}, got.Leader)
}

func TestDerive_SingleInstance(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	got := leader.Derive(instances(instance("conn-a", 100, electionMode)))

	require.True(got.HasElection)
	require.True(got.Found)
	require.Equal("conn-a", got.Leader.WsConnectionId)
}

func TestDerive_NoInstances(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	got := leader.Derive(nil)

	require.False(got.HasElection)
	require.False(got.Found)
	require.Equal(entity.Backend{}, got.Leader)
}

func TestDeriveByModule(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	first := instance("conn-a", 100, electionMode)
	first.ModuleId = "module-1"
	first.ModuleName = "module-one"

	second := instance("conn-b", 200, noElectionMode)
	second.ModuleId = "module-2"
	second.ModuleName = "module-two"

	elections := leader.DeriveByModule(instances(second, first))

	require.Len(elections, 2, "every module gets its own election")
	require.Equal("conn-a", elections["module-1"].Leader.WsConnectionId)
	require.False(elections["module-2"].HasElection, "an instance of another module must not start an election")
}

func instance(wsConnectionId string, createdAt int64, mode string) entity.Backend {
	return entity.Backend{
		WsConnectionId:   wsConnectionId,
		ModuleId:         "module-id",
		ModuleName:       "module",
		Address:          "10.2.9.1:9999",
		ModuleLeaderMode: mode,
		CreatedAt:        xtypes.Time(time.Unix(createdAt, 0)),
	}
}

func instances(backends ...entity.Backend) []entity.Backend {
	return backends
}

func permutations(backends []entity.Backend) [][]entity.Backend {
	reversed := slices.Clone(backends)
	slices.Reverse(reversed)

	rotated := append(slices.Clone(backends[1:]), backends[0])

	return [][]entity.Backend{backends, reversed, rotated}
}
