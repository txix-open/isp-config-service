package tests_test

import (
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/rqlite/db"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
)

const (
	recomputeModule = "leaderrecompute"
	recomputeHost   = "10.2.9.21"
)

// nolint:tagliatelle
type recomputeTotalRow struct {
	Total int `json:"total"`
}

// nolint:tagliatelle
type recomputeModuleRow struct {
	Id string `json:"id"`
}

//nolint:paralleltest
func TestModuleLeaderRecomputeEventsInQueue(t *testing.T) {
	require, logger, testDb := setupTestApp(t)

	instance := startModuleInstance(t, recomputeModule, recomputeHost, cluster.ModuleLeaderModeRequestElection, logger)

	moduleId := requireRecomputeModuleId(t, testDb, recomputeModule)

	require.Eventually(func() bool {
		return countLeaderRecomputeEvents(t, testDb, moduleId) == 1
	}, announceTimeout, announceTick, "a connected instance must add exactly one leader recompute event")
	require.Equal(1, countLeaderRecomputeEventsByReason(t, testDb, moduleId, entity.ModuleLeaderReasonReady))

	instance.disconnect(t)

	require.Eventually(func() bool {
		return countLeaderRecomputeEvents(t, testDb, moduleId) == 2
	}, announceTimeout, announceTick, "a disconnected instance must add exactly one leader recompute event")
	require.Equal(1, countLeaderRecomputeEventsByReason(t, testDb, moduleId, entity.ModuleLeaderReasonDisconnect))
}

func requireRecomputeModuleId(t *testing.T, testDb db.DB, moduleName string) string {
	t.Helper()
	require := require.New(t)

	var row recomputeModuleRow
	require.Eventually(func() bool {
		found := recomputeModuleRow{}
		err := testDb.SelectRow(t.Context(), &found, `
			select id
			from isp_config_service__module
			where name = ?`, moduleName)
		if err != nil {
			return false
		}
		row = found
		return row.Id != ""
	}, announceTimeout, announceTick, "module %s must be registered", moduleName)

	return row.Id
}

func countLeaderRecomputeEvents(t *testing.T, testDb db.DB, moduleId string) int {
	t.Helper()
	require := require.New(t)

	row := recomputeTotalRow{}
	err := testDb.SelectRow(t.Context(), &row, `
		select count(*) as total
		from isp_config_service__event
		where cast(payload as text) like ?
		  and cast(payload as text) like ?`,
		"%moduleLeaderRecompute%", moduleIdLikePattern(moduleId))
	require.NoError(err)

	return row.Total
}

func countLeaderRecomputeEventsByReason(t *testing.T, testDb db.DB, moduleId string, reason string) int {
	t.Helper()
	require := require.New(t)

	row := recomputeTotalRow{}
	err := testDb.SelectRow(t.Context(), &row, `
		select count(*) as total
		from isp_config_service__event
		where cast(payload as text) like ?
		  and cast(payload as text) like ?
		  and cast(payload as text) like ?`,
		"%moduleLeaderRecompute%", moduleIdLikePattern(moduleId), "%\"reason\":\""+reason+"\"%")
	require.NoError(err)

	return row.Total
}

func moduleIdLikePattern(moduleId string) string {
	return "%\"moduleId\":\"" + moduleId + "\"%"
}
