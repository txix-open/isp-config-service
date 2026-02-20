package tests_test

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"isp-config-service/conf"
	"isp-config-service/domain"
	"isp-config-service/service/startup"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/test/fake"
)

// nolint:paralleltest
func TestMetricsAutodiscovery(t *testing.T) {
	require, logger := setupTestWithAppRun(t)

	cli := httpclix.DefaultWithBalancer([]string{"127.0.0.1:9006"})
	resp := make(domain.AutodiscoveryResponse, 0)
	err := cli.Get("/internal/metrics/autodiscovery").
		JsonResponseBody(&resp).
		StatusCodeToError().
		DoWithoutResponse(t.Context())
	require.NoError(err)

	configTarget := domain.PrometheusTargets{
		Targets: []string{"isp-config-service:9006"},
		Labels: map[string]string{
			"app":              "isp-config-service",
			"__metrics_path__": "/internal/metrics",
		},
	}
	require.Equal(domain.AutodiscoveryResponse{configTarget}, resp)

	const (
		iterCount         = 8
		chanceToMetricsAd = 0.5
	)
	expected := make(domain.AutodiscoveryResponse, 0, iterCount+1)
	for i := range iterCount {
		metricsAd := generateMetricsAdWithChance(chanceToMetricsAd)
		if metricsAd != nil {
			expected = append(expected, domain.PrometheusTargets{
				Targets: []string{metricsAd.Address},
				Labels:  metricsAd.Labels,
			})
		}
		clusterCli := newClusterClientWith(
			t,
			strconv.Itoa(i),
			"10.2.9.1",
			testModuleRemoteConfig{},
			[]byte("{}"),
			metricsAd,
			logger,
		)
		go func() {
			err := clusterCli.Run(t.Context(), cluster.NewEventHandler())
			require.NoError(err) // nolint:testifylint
		}()
	}
	time.Sleep(time.Second)
	expected = append(expected, configTarget)

	resp = make(domain.AutodiscoveryResponse, 0)
	err = cli.Get("/internal/metrics/autodiscovery").
		JsonResponseBody(&resp).
		StatusCodeToError().
		DoWithoutResponse(t.Context())
	require.NoError(err)
	require.Equal(expected, resp)
}

func generateMetricsAdWithChance(chance float64) *cluster.MetricsAutodiscovery {
	if rand.Float64() > chance { // nolint:gosec
		return nil
	}
	return fake.It[*cluster.MetricsAutodiscovery]()
}

// nolint:ireturn
func setupTestWithAppRun(t *testing.T) (*require.Assertions, log.Logger) {
	t.Helper()
	require := require.New(t)

	t.Setenv("APP_CONFIG_PATH", "../conf/config.yml")
	t.Setenv("DefaultRemoteConfigPath", "../conf/default_remote_config.json")
	boot := bootstrap.New("1.0.0", conf.Remote{}, nil, cluster.GrpcTransport)
	boot.MigrationsDir = "../migrations"
	dataPath := dataDir(t)
	boot.App.Config().Set("rqlite.DataPath", dataPath)
	boot.App.Config().Set("backup.Enable", false)
	logger := boot.App.Logger()

	startup, err := startup.New(boot)
	require.NoError(err)
	t.Cleanup(func() {
		for _, closer := range startup.Closers() {
			err := closer.Close()
			require.NoError(err)
		}
	})
	err = startup.Run(t.Context())
	require.NoError(err)

	go func() {
		err = boot.App.Run()
		require.NoError(err) // nolint:testifylint
	}()

	time.Sleep(1 * time.Second)

	return require, logger
}
