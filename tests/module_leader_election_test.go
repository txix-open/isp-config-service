package tests_test

import (
	"context"
	"net"
	"slices"
	"strings"
	"testing"

	"isp-config-service/service/rqlite/db"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
)

const (
	electionHostOne   = "10.2.9.1"
	electionHostTwo   = "10.2.9.2"
	electionHostThree = "10.2.9.3"

	plainModule         = "plain"
	plainHostOne        = "10.2.9.7"
	plainHostTwo        = "10.2.9.8"
	plainObserverModule = "plainobserver"
	plainObserverIP     = "10.2.9.10"

	electionObserverModule = "electionobserver"
	electionObserverIP     = "10.2.9.9"

	lateObserverModule = "lateobserver"
	lateObserverIP     = "10.2.9.11"

	noLeaderModule       = "noleader"
	noLeaderHostOne      = "10.2.9.12"
	noLeaderHostTwo      = "10.2.9.13"
	noLeaderObserverIP   = "10.2.9.14"
	noLeaderObserverName = "noleaderobserver"

	instancePort = "9999"
)

// nolint:tagliatelle
type leaderOrderRow struct {
	WsConnectionId string `json:"ws_connection_id"`
	Address        string `json:"address"`
}

// nolint:tagliatelle
type schemaNameRow struct {
	Name string `json:"name"`
}

type moduleInstance struct {
	host   string
	cli    *cluster.Client
	cancel context.CancelFunc
}

//nolint:funlen,paralleltest
func TestModuleLeaderElection(t *testing.T) {
	require, logger, testDb := setupTestApp(t)
	ctx := t.Context()

	electionObserver := newClusterEventHandler()
	runClusterClient(ctx, require,
		newClusterClient(t, electionObserverModule, electionObserverIP, logger),
		cluster.NewEventHandler().RequireModule(electionModule, electionObserver).RoutesReceiver(electionObserver),
	)

	plainObserver := newClusterEventHandler()
	runClusterClient(ctx, require,
		newClusterClient(t, plainObserverModule, plainObserverIP, logger),
		cluster.NewEventHandler().RequireModule(plainModule, plainObserver).RoutesReceiver(plainObserver),
	)

	electedInstances := []moduleInstance{
		startModuleInstance(t, electionModule, electionHostOne, cluster.ModuleLeaderModeRequestElection, logger),
		startModuleInstance(t, electionModule, electionHostTwo, cluster.ModuleLeaderModeRequestElection, logger),
		startModuleInstance(t, electionModule, electionHostThree, cluster.ModuleLeaderModeRequestElection, logger),
	}

	byHost := make(map[string]moduleInstance, len(electedInstances))
	for _, instance := range electedInstances {
		byHost[instance.host] = instance
	}

	startModuleInstance(t, plainModule, plainHostOne, cluster.ModuleLeaderModeNone, logger)
	startModuleInstance(t, plainModule, plainHostTwo, cluster.ModuleLeaderModeNone, logger)

	announced := []string{addressOf(electionHostOne), addressOf(electionHostTwo), addressOf(electionHostThree)}

	require.Eventually(func() bool {
		return len(selectLeaderOrder(ctx, testDb, electionModule)) == 3
	}, announceTimeout, announceTick, "all three instances of %s must be connected", electionModule)

	elected := selectLeaderOrder(ctx, testDb, electionModule)
	expectedLeader := elected[0].Address
	expectedSuccessor := elected[1].Address

	require.Eventually(func() bool {
		return slices.Equal(lastHosts(electionObserver), []string{expectedLeader})
	}, announceTimeout, announceTick, "the address list of an elected module must keep exactly the derived leader")

	received := lastHosts(electionObserver)
	require.Len(received, 1, "an elected module must be balanced to a single address")
	require.Contains(announced, received[0], "the leader must be one of the announced instances")

	require.Eventually(func() bool {
		return routingLeaderOf(electionObserver, electionModule) == expectedLeader
	}, announceTimeout, announceTick, "the routing map of an elected module must publish only the leader")
	requireRoutingLeader(require, lastRoutesOf(electionObserver, electionModule), expectedLeader)

	require.Eventually(func() bool {
		return len(lastRoutesOf(electionObserver, plainModule)) == 2
	}, announceTimeout, announceTick, "a module without election must publish every instance")
	requireRoutingUntouched(require, lastRoutesOf(electionObserver, plainModule), plainHosts())

	lateObserver := newClusterEventHandler()
	runClusterClient(ctx, require,
		newClusterClient(t, lateObserverModule, lateObserverIP, logger),
		cluster.NewEventHandler().RequireModule(electionModule, lateObserver).RoutesReceiver(lateObserver),
	)

	require.Eventually(func() bool {
		return slices.Equal(lastHosts(lateObserver), []string{expectedLeader}) &&
			routingLeaderOf(lateObserver, electionModule) == expectedLeader
	}, announceTimeout, announceTick, "a subscriber connected after the election must see the same leader")
	requireRoutingLeader(require, lastRoutesOf(lateObserver, electionModule), expectedLeader)

	standbySnapshots := len(electionObserver.ReceivedHosts())
	byHost[hostOf(t, elected[2].Address)].disconnect(t)

	require.Eventually(func() bool {
		return len(electionObserver.ReceivedHosts()) > standbySnapshots
	}, announceTimeout, announceTick, "disconnecting a standby must re-emit the address list")
	require.Equal([]string{expectedLeader}, lastHosts(electionObserver),
		"disconnecting a standby must not move leadership")

	leaderSnapshots := len(electionObserver.ReceivedHosts())
	byHost[hostOf(t, expectedLeader)].disconnect(t)

	require.Eventually(func() bool {
		return slices.Equal(lastHosts(electionObserver), []string{expectedSuccessor})
	}, announceTimeout, announceTick, "disconnecting the leader must move leadership to the next one in derivation order")
	require.Equal(expectedSuccessor, selectLeaderOrder(ctx, testDb, electionModule)[0].Address)
	requireRoutingLeader(require, lastRoutesOf(electionObserver, electionModule), expectedSuccessor)

	require.Greater(len(electionObserver.ReceivedHosts()), leaderSnapshots,
		"disconnecting the leader must re-emit the address list")

	lastSnapshots := len(electionObserver.ReceivedHosts())
	byHost[hostOf(t, expectedSuccessor)].disconnect(t)

	require.Eventually(func() bool {
		return len(electionObserver.ReceivedHosts()) > lastSnapshots
	}, announceTimeout, announceTick, "disconnecting the last instance must re-emit the address list")
	require.Empty(lastHosts(electionObserver), "with no instances left the subscriber must receive an empty list")
	require.Empty(selectLeaderOrder(ctx, testDb, electionModule))

	require.Eventually(func() bool {
		return slices.Equal(lastHosts(plainObserver), selectEmittedAddresses(ctx, testDb, plainModule))
	}, announceTimeout, announceTick, "a module without election keeps its newest-first address list")
	requireRoutingUntouched(require, lastRoutesOf(plainObserver, plainModule), plainHosts())

	requireNoLeaderState(t, testDb)
}

//nolint:paralleltest
func TestModuleLeaderElectionWithoutLeader(t *testing.T) {
	require, logger, testDb := setupTestApp(t)
	ctx := t.Context()

	observer := newClusterEventHandler()
	runClusterClient(ctx, require,
		newClusterClient(t, noLeaderObserverName, noLeaderObserverIP, logger),
		cluster.NewEventHandler().RequireModule(noLeaderModule, observer).RoutesReceiver(observer),
	)

	byHost := map[string]moduleInstance{
		noLeaderHostOne: startModuleInstance(t, noLeaderModule, noLeaderHostOne, cluster.ModuleLeaderModeRequestElection, logger),
		noLeaderHostTwo: startModuleInstance(t, noLeaderModule, noLeaderHostTwo, cluster.ModuleLeaderModeRequestElection, logger),
	}

	require.Eventually(func() bool {
		return len(selectLeaderOrder(ctx, testDb, noLeaderModule)) == 2
	}, announceTimeout, announceTick, "both instances of %s must be connected", noLeaderModule)

	order := selectLeaderOrder(ctx, testDb, noLeaderModule)
	require.Eventually(func() bool {
		return slices.Equal(lastHosts(observer), []string{order[0].Address})
	}, announceTimeout, announceTick, "the leader address must be the only published address")
	requireRoutingLeader(require, lastRoutesOf(observer, noLeaderModule), order[0].Address)

	byHost[hostOf(t, order[0].Address)].disconnect(t)

	var remaining []leaderOrderRow
	require.Eventually(func() bool {
		remaining = selectLeaderOrder(ctx, testDb, noLeaderModule)
		return len(remaining) == 1 && slices.Equal(lastHosts(observer), []string{remaining[0].Address})
	}, announceTimeout, announceTick, "leadership must move to the remaining instance")

	snapshots := len(observer.ReceivedHosts())
	byHost[hostOf(t, remaining[0].Address)].disconnect(t)

	require.Eventually(func() bool {
		return len(observer.ReceivedHosts()) > snapshots && len(lastHosts(observer)) == 0
	}, announceTimeout, announceTick, "with no instances left the subscriber must receive an empty address list")
	require.Empty(lastHosts(observer), "the client must receive an empty list, not the previous address")
	require.Empty(selectLeaderOrder(ctx, testDb, noLeaderModule))
	require.Empty(lastRoutesOf(observer, noLeaderModule), "a module without instances publishes no declaration")
}

func startModuleInstance(
	t *testing.T,
	moduleName string,
	host string,
	moduleLeaderMode string,
	logger log.Logger,
) moduleInstance {
	t.Helper()

	cli := newClusterClientFull(t, moduleName, host, testModuleRemoteConfig{}, []byte("{}"), nil, moduleLeaderMode, logger)
	ctx, cancel := context.WithCancel(t.Context())
	runClusterClient(ctx, require.New(t), cli, cluster.NewEventHandler())

	return moduleInstance{host: host, cli: cli, cancel: cancel}
}

func (i moduleInstance) disconnect(t *testing.T) {
	t.Helper()

	i.cancel()
	require.NoError(t, i.cli.Close())
}

func addressOf(host string) string {
	return net.JoinHostPort(host, instancePort)
}

func hostOf(t *testing.T, address string) string {
	t.Helper()

	host, _, err := net.SplitHostPort(address)
	require.New(t).NoError(err)

	return host
}

func plainHosts() []string {
	return []string{addressOf(plainHostOne), addressOf(plainHostTwo)}
}

func lastHosts(handler *clusterEventHandler) []string {
	snapshots := handler.ReceivedHosts()
	if len(snapshots) == 0 {
		return nil
	}

	return snapshots[len(snapshots)-1]
}

func lastRoutesOf(handler *clusterEventHandler, moduleName string) []cluster.BackendDeclaration {
	snapshots := handler.ReceivedRoutes()
	if len(snapshots) == 0 {
		return nil
	}

	declarations := make([]cluster.BackendDeclaration, 0)
	for _, declaration := range snapshots[len(snapshots)-1] {
		if declaration.ModuleName == moduleName {
			declarations = append(declarations, declaration)
		}
	}

	return declarations
}

func routingLeaderOf(handler *clusterEventHandler, moduleName string) string {
	declarations := lastRoutesOf(handler, moduleName)
	if len(declarations) != 1 || !declarations[0].Address.IsModuleLeader {
		return ""
	}

	return addressOf(declarations[0].Address.IP)
}

func requireRoutingLeader(require *require.Assertions, declarations []cluster.BackendDeclaration, leaderAddress string) {
	require.Len(declarations, 1, "the routing map of an elected module must publish only the leader")
	require.Equal(leaderAddress, addressOf(declarations[0].Address.IP))
	require.True(declarations[0].Address.IsModuleLeader, "the leader declaration must carry the leader flag")
}

func requireRoutingUntouched(require *require.Assertions, declarations []cluster.BackendDeclaration, announced []string) {
	require.Len(declarations, len(announced), "a module without election must publish every instance")

	addresses := make([]string, 0, len(declarations))
	for _, declaration := range declarations {
		require.False(declaration.Address.IsModuleLeader)
		addresses = append(addresses, addressOf(declaration.Address.IP))
	}
	require.ElementsMatch(announced, addresses)
}

func selectLeaderOrder(ctx context.Context, testDb db.DB, moduleName string) []leaderOrderRow {
	rows := make([]leaderOrderRow, 0)
	_ = testDb.Select(ctx, &rows, `
		select ws_connection_id, address
		from isp_config_service__backend
		where module_name = ?
		order by created_at asc, ws_connection_id asc`, moduleName)

	return rows
}

func selectEmittedAddresses(ctx context.Context, testDb db.DB, moduleName string) []string {
	rows := selectLeaderOrder(ctx, testDb, moduleName)
	addresses := make([]string, 0, len(rows))
	for i := range slices.Backward(rows) {
		addresses = append(addresses, rows[i].Address)
	}

	return addresses
}

func requireNoLeaderState(t *testing.T, testDb db.DB) {
	t.Helper()
	require := require.New(t)

	columns := make([]schemaNameRow, 0)
	err := testDb.Select(t.Context(), &columns, `pragma table_info('isp_config_service__backend')`)
	require.NoError(err)

	leaderColumns := make([]string, 0)
	for _, column := range columns {
		if strings.Contains(strings.ToLower(column.Name), "leader") {
			leaderColumns = append(leaderColumns, column.Name)
		}
	}
	require.Equal([]string{"module_leader_mode"}, leaderColumns,
		"only the announced mode is stored, the derived leader never is")

	objects := make([]schemaNameRow, 0)
	err = testDb.Select(t.Context(), &objects, `
		select name from sqlite_master
		where lower(name) like '%leader%' and type in ('table','view','index','trigger')`)
	require.NoError(err)
	require.Empty(objects, "leader election must not add a table, view, index, or trigger")
}
