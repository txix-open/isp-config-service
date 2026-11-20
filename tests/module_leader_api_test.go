package tests_test

import (
	"context"
	"net"
	"slices"
	"testing"

	"isp-config-service/domain"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
)

const (
	apiObserverModule = "apiobserver"
	apiObserverIP     = "10.2.9.20"

	apiSilentModule = "apisilent"
	apiSilentHost   = "10.2.9.21"

	modulesInfoEndpoint = "config/module/get_modules_info"
	routesEndpoint      = "config/routing/get_routes"
)

//nolint:funlen,paralleltest
func TestModuleLeaderApiFlag(t *testing.T) {
	require, logger, testDb := setupTestApp(t)
	ctx := t.Context()

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})

	observer := newClusterEventHandler()
	runClusterClient(ctx, require,
		newClusterClient(t, apiObserverModule, apiObserverIP, logger),
		cluster.NewEventHandler().RequireModule(electionModule, observer).RoutesReceiver(observer),
	)

	instances := []moduleInstance{
		startModuleInstance(t, electionModule, electionHostOne, cluster.ModuleLeaderModeRequestElection, logger),
		startModuleInstance(t, electionModule, electionHostTwo, cluster.ModuleLeaderModeRequestElection, logger),
		startModuleInstance(t, electionModule, electionHostThree, cluster.ModuleLeaderModeRequestElection, logger),
		startModuleInstance(t, plainModule, plainHostOne, cluster.ModuleLeaderModeNone, logger),
		startModuleInstance(t, plainModule, plainHostTwo, cluster.ModuleLeaderModeNone, logger),
		startModuleInstance(t, apiSilentModule, apiSilentHost, cluster.ModuleLeaderModeNone, logger),
	}

	byHost := make(map[string]moduleInstance, len(instances))
	for _, instance := range instances {
		byHost[instance.host] = instance
	}

	connected := map[string]int{electionModule: 3, plainModule: 2, apiSilentModule: 1}
	require.Eventually(func() bool {
		return len(selectLeaderOrder(ctx, testDb, electionModule)) == 3 &&
			len(selectLeaderOrder(ctx, testDb, plainModule)) == 2 &&
			len(selectLeaderOrder(ctx, testDb, apiSilentModule)) == 1
	}, announceTimeout, announceTick, "every instance must be connected")

	elected := selectLeaderOrder(ctx, testDb, electionModule)
	derivedLeader := elected[0].Address

	require.Eventually(func() bool {
		return slices.Equal(lastHosts(observer), []string{derivedLeader})
	}, announceTimeout, announceTick, "the subscriber must be balanced to the derived leader")
	observedLeader := lastHosts(observer)[0]

	status := requireApiCounts(t, ctx, apiCli, modulesInfoEndpoint, instanceCount, connected)

	electionInfo := findModule(status, electionModule)
	require.NotNil(electionInfo, electionModule)
	require.True(electionInfo.Active)
	require.Len(electionInfo.Status, 3, "modules info must keep every instance of the elected module")

	apiLeader := requireSingleLeader(require, electionInfo.Status)
	require.Equal(derivedLeader, apiLeader, "the leader of the API must be the leader derived from the DB")
	require.Equal(observedLeader, apiLeader, "the leader of the API must be the address the subscriber receives")

	requireNoLeaders(require, connectionsOf(status, plainModule), plainModule)
	requireNoLeaders(require, connectionsOf(status, apiSilentModule), apiSilentModule)

	routes := requireApiCounts(t, ctx, apiCli, routesEndpoint, connectionCount, connected)
	require.Len(connectionsOfModule(routes, electionModule), 3, "routes must not narrow the elected module")
	require.Equal(apiLeader, requireSingleLeader(require, connectionsOfModule(routes, electionModule)),
		"routes must mark the same leader as modules info")
	requireNoLeaders(require, connectionsOfModule(routes, plainModule), plainModule)

	byHost[apiSilentHost].disconnect(t)

	byHost[hostOf(t, elected[1].Address)].disconnect(t)
	byHost[hostOf(t, elected[2].Address)].disconnect(t)

	status = requireApiCounts(t, ctx, apiCli, modulesInfoEndpoint, instanceCount, map[string]int{
		electionModule:  1,
		plainModule:     2,
		apiSilentModule: 0,
	})

	silentInfo := findModule(status, apiSilentModule)
	require.NotNil(silentInfo, apiSilentModule)
	require.False(silentInfo.Active)
	require.Empty(silentInfo.Status, "a module without instances has no connection to mark")

	electionInfo = findModule(status, electionModule)
	require.NotNil(electionInfo, electionModule)
	require.True(electionInfo.Active)
	require.Len(electionInfo.Status, 1)
	require.True(electionInfo.Status[0].IsModuleLeader, "the only instance of an elected module is its leader")
	require.Equal(derivedLeader, connectionAddress(electionInfo.Status[0]),
		"dropping standbys must not move leadership")
}

func connectionsOf(modules []domain.ModuleInfo, moduleName string) []domain.Connection {
	for _, module := range modules {
		if module.Name == moduleName {
			return module.Status
		}
	}

	return nil
}

func connectionsOfModule(connections []domain.Connection, moduleName string) []domain.Connection {
	ofModule := make([]domain.Connection, 0, len(connections))
	for _, connection := range connections {
		if connection.ModuleName == moduleName {
			ofModule = append(ofModule, connection)
		}
	}

	return ofModule
}

func connectionAddress(connection domain.Connection) string {
	return net.JoinHostPort(connection.Address.Ip, connection.Address.Port)
}

func requireSingleLeader(require *require.Assertions, connections []domain.Connection) string {
	leaders := make([]string, 0, 1)
	for _, connection := range connections {
		if connection.IsModuleLeader {
			leaders = append(leaders, connectionAddress(connection))
		}
	}
	require.Len(leaders, 1, "an elected module must expose exactly one leader")

	return leaders[0]
}

func requireNoLeaders(require *require.Assertions, connections []domain.Connection, moduleName string) {
	require.NotEmpty(connections, moduleName)
	for _, connection := range connections {
		require.False(connection.IsModuleLeader, "a module without election must not expose a leader: "+moduleName)
	}
}

func instanceCount(module domain.ModuleInfo) map[string]int {
	return map[string]int{module.Name: len(module.Status)}
}

func connectionCount(connection domain.Connection) map[string]int {
	return map[string]int{connection.ModuleName: 1}
}

func requireApiCounts[T any](
	t *testing.T,
	ctx context.Context,
	apiCli *client.Client,
	endpoint string,
	countsOf func(row T) map[string]int,
	expected map[string]int,
) []T {
	t.Helper()
	require := require.New(t)

	rows := make([]T, 0)
	require.Eventually(func() bool {
		latest := make([]T, 0)
		err := apiCli.Invoke(endpoint).JsonResponseBody(&latest).Do(ctx)
		if err != nil {
			return false
		}
		rows = latest

		return countsMatch(countsByModule(rows, countsOf), expected)
	}, announceTimeout, announceTick, endpoint+" must report every instance")

	return rows
}

func countsByModule[T any](rows []T, countsOf func(row T) map[string]int) map[string]int {
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		for moduleName, total := range countsOf(row) {
			counts[moduleName] += total
		}
	}

	return counts
}

func countsMatch(counts, expected map[string]int) bool {
	for moduleName, total := range expected {
		if counts[moduleName] != total {
			return false
		}
	}

	return true
}
