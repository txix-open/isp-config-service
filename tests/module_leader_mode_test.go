package tests_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"isp-config-service/domain"
	"isp-config-service/service/rqlite/db"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/log"
)

const (
	electionModule = "election"
	noneModule     = "none"
	leaderModule   = "imleader"
	followerModule = "imfollower"
	garbageModule  = "garbage"
	observerModule = "observer"

	garbageLeaderMode = "SOMETHING_ELSE"

	announceTimeout = 20 * time.Second
	announceTick    = 200 * time.Millisecond
)

// nolint:tagliatelle
type backendRow struct {
	WsConnectionId   string  `json:"ws_connection_id"`
	CreatedAt        int64   `json:"created_at"`
	ModuleLeaderMode *string `json:"module_leader_mode"`
}

// nolint:tagliatelle
type backendCountRow struct {
	Total int `json:"total"`
}

// nolint:tagliatelle
type epochAdvancedRow struct {
	Advanced int `json:"advanced"`
}

func (r backendRow) moduleLeaderMode() string {
	if r.ModuleLeaderMode == nil {
		return "null"
	}
	return *r.ModuleLeaderMode
}

//nolint:funlen,paralleltest
func TestModuleLeaderMode(t *testing.T) {
	require, logger, testDb := setupTestApp(t)
	ctx := t.Context()

	routes := newClusterEventHandler()
	runClusterClient(ctx, require,
		newClusterClient(t, observerModule, "10.2.9.9", logger),
		cluster.NewEventHandler().RoutesReceiver(routes),
	)

	electionCli := newClusterClientWithLeaderMode(t, electionModule, "10.2.9.1", cluster.ModuleLeaderModeRequestElection, logger)
	runClusterClient(ctx, require, electionCli, cluster.NewEventHandler())

	noneCli := newClusterClient(t, noneModule, "10.2.9.2", logger)
	runClusterClient(ctx, require, noneCli, cluster.NewEventHandler())

	leaderCli := newClusterClientWithLeaderMode(t, leaderModule, "10.2.9.3", cluster.ModuleLeaderModeImLeader, logger)
	runClusterClient(ctx, require, leaderCli, cluster.NewEventHandler())

	followerCli := newClusterClientWithLeaderMode(t, followerModule, "10.2.9.4", cluster.ModuleLeaderModeImFollower, logger)
	runClusterClient(ctx, require, followerCli, cluster.NewEventHandler())

	garbageCli := newClusterClientWithLeaderMode(t, garbageModule, "10.2.9.5", garbageLeaderMode, logger)
	runClusterClient(ctx, require, garbageCli, cluster.NewEventHandler())

	electionRow := requireBackendRow(t, testDb, electionModule)
	require.Equal(cluster.ModuleLeaderModeRequestElection, electionRow.moduleLeaderMode())

	noneRow := requireBackendRow(t, testDb, noneModule)
	require.NotNil(noneRow.ModuleLeaderMode, "mode announced as empty must be stored as empty string")
	require.Equal(cluster.ModuleLeaderModeNone, noneRow.moduleLeaderMode())
	t.Logf("rqlite returns module_leader_mode of %s as %q", noneModule, noneRow.moduleLeaderMode())

	reserved := []struct {
		moduleName string
		cli        *cluster.Client
	}{
		{moduleName: leaderModule, cli: leaderCli},
		{moduleName: followerModule, cli: followerCli},
		{moduleName: garbageModule, cli: garbageCli},
	}
	for _, instance := range reserved {
		row := requireBackendRow(t, testDb, instance.moduleName)
		require.NotNil(row.ModuleLeaderMode, "reserved mode must be stored as empty string")
		require.Equal(cluster.ModuleLeaderModeNone, row.moduleLeaderMode(), instance.moduleName)

		require.Eventually(func() bool {
			return instance.cli.Healthcheck(ctx) == nil
		}, announceTimeout, announceTick, "handshake of %s must not be rejected", instance.moduleName)
	}

	requireAcceptedModules(require, ctx, leaderModule, followerModule, garbageModule)

	for _, routing := range requireReceivedRoutes(require, routes) {
		for _, declaration := range routing {
			// electionModule declared REQUEST_ELECTION: its leader is published with the leader flag set.
			if declaration.ModuleName == electionModule {
				continue
			}

			require.False(declaration.Address.IsModuleLeader)
			require.Equal(cluster.ModuleLeaderModeNone, declaration.ModuleLeaderMode)
		}
	}

	require.Eventually(func() bool {
		return backendCreatedAtPassed(t, testDb, electionModule)
	}, announceTimeout, announceTick, "created_at must be older than now to detect a move")

	beforeReannounce, err := selectBackendRow(t, testDb, electionModule)
	require.NoError(err)

	err = electionCli.SendModuleReady(ctx)
	require.NoError(err)

	reannouncedRow, err := selectBackendRow(t, testDb, electionModule)
	require.NoError(err)
	require.Equal(beforeReannounce.WsConnectionId, reannouncedRow.WsConnectionId)
	require.Equal(beforeReannounce.CreatedAt, reannouncedRow.CreatedAt, "re-announcement must not move the instance")
	require.Equal(cluster.ModuleLeaderModeRequestElection, reannouncedRow.moduleLeaderMode())

	count := selectBackendCount(t, testDb, electionModule)
	require.Equal(1, count.Total)
}

func newClusterClientWithLeaderMode(
	t *testing.T,
	moduleName string,
	host string,
	moduleLeaderMode string,
	logger log.Logger,
) *cluster.Client {
	t.Helper()
	return newClusterClientFull(
		t,
		moduleName,
		host,
		testModuleRemoteConfig{},
		[]byte("{}"),
		nil,
		moduleLeaderMode,
		logger,
	)
}

func runClusterClient(
	ctx context.Context,
	require *require.Assertions,
	cli *cluster.Client,
	handler *cluster.EventHandler,
) {
	go func() {
		err := cli.Run(ctx, handler)
		require.NoError(err) //nolint:testifylint
	}()
}

func selectBackendRow(t *testing.T, testDb db.DB, moduleName string) (backendRow, error) {
	t.Helper()

	var row backendRow
	err := testDb.SelectRow(t.Context(), &row, `
		select ws_connection_id, created_at, module_leader_mode
		from isp_config_service__backend
		where module_name = ?`, moduleName)
	if err != nil {
		return backendRow{}, errors.WithMessagef(err, "select backend row of module %s", moduleName)
	}

	return row, nil
}

func requireBackendRow(t *testing.T, testDb db.DB, moduleName string) backendRow {
	t.Helper()
	require := require.New(t)

	var row backendRow
	require.Eventually(func() bool {
		stored, err := selectBackendRow(t, testDb, moduleName)
		if err != nil {
			return false
		}
		row = stored
		return true
	}, announceTimeout, announceTick, "module %s announcement is not stored", moduleName)

	return row
}

func backendCreatedAtPassed(t *testing.T, testDb db.DB, moduleName string) bool {
	t.Helper()

	var row epochAdvancedRow
	err := testDb.SelectRow(t.Context(), &row, `
		select unixepoch() > created_at as advanced
		from isp_config_service__backend
		where module_name = ?`, moduleName)

	return err == nil && row.Advanced == 1
}

func selectBackendCount(t *testing.T, testDb db.DB, moduleName string) backendCountRow {
	t.Helper()

	row := backendCountRow{}
	err := testDb.SelectRow(t.Context(), &row, `
		select count(*) as total
		from isp_config_service__backend
		where module_name = ?`, moduleName)
	require.New(t).NoError(err)

	return row
}

func requireAcceptedModules(require *require.Assertions, ctx context.Context, moduleNames ...string) {
	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})

	modules := make([]domain.ModuleInfo, 0)
	require.Eventually(func() bool {
		modules = modules[:0]
		err = apiCli.Invoke("config/module/get_modules_info").
			JsonResponseBody(&modules).
			Do(ctx)
		if err != nil {
			return false
		}

		connected := make([]string, 0, len(modules))
		for _, module := range modules {
			connected = append(connected, module.Name)
		}
		for _, moduleName := range moduleNames {
			if !slices.Contains(connected, moduleName) {
				return false
			}
		}
		return true
	}, announceTimeout, announceTick, "announced modules must be visible in modules info")

	for _, moduleName := range moduleNames {
		module := findModule(modules, moduleName)
		require.NotNil(module, moduleName)
		require.Len(module.Status, 1, moduleName)
	}
}

func findModule(modules []domain.ModuleInfo, moduleName string) *domain.ModuleInfo {
	for i := range modules {
		if modules[i].Name == moduleName {
			return &modules[i]
		}
	}
	return nil
}

func requireReceivedRoutes(require *require.Assertions, handler *clusterEventHandler) []cluster.RoutingConfig {
	require.Eventually(func() bool {
		return len(handler.ReceivedRoutes()) > 0
	}, announceTimeout, announceTick, "routing configuration must be received")

	return handler.ReceivedRoutes()
}
