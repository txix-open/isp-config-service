package tests_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/rc/schema"
	"isp-config-service/conf"
	"isp-config-service/domain"
	"isp-config-service/service/startup"
)

//nolint:funlen
func TestAcceptance(t *testing.T) {
	require := require.New(t)

	t.Setenv("APP_CONFIG_PATH", "../conf/config.yml")
	t.Setenv("DefaultRemoteConfigPath", "../conf/default_remote_config.json")
	boot := bootstrap.New("1.0.0", conf.Remote{}, nil)
	boot.MigrationsDir = "../migrations"
	dataPath := dataDir(t)
	boot.App.Config().Set("rqlite.DataPath", dataPath)
	logger := boot.App.Logger()

	startup, err := startup.New(boot)
	require.NoError(err)
	t.Cleanup(func() {
		for _, closer := range startup.Closers() {
			err := closer.Close()
			require.NoError(err)
		}
	})
	err = startup.Run(context.Background())
	require.NoError(err)
	time.Sleep(1 * time.Second)

	clientA1 := newClusterClient("A", "10.2.9.1", logger)
	go func() {
		handler := cluster.NewEventHandler()
		err := clientA1.Run(context.Background(), handler)
		require.NoError(err) //nolint:testifylint
	}()

	clientA2 := newClusterClient("A", "10.2.9.2", logger)
	clientA2Ctx, cancelClient2 := context.WithCancel(context.Background())
	go func() {
		handler := cluster.NewEventHandler()
		err := clientA2.Run(clientA2Ctx, handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(2 * time.Second)

	clientB := newClusterClient("B", "10.2.9.2", logger)
	eventHandler := newClusterEventHandler()
	go func() {
		handler := cluster.NewEventHandler().
			RemoteConfigReceiver(eventHandler).
			RequireModule("A", eventHandler).
			RoutesReceiver(eventHandler)
		err := clientB.Run(context.Background(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(2 * time.Second)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})

	activeConfig := domain.Config{}
	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "B"}).
		JsonResponseBody(&activeConfig).
		Do(context.Background())
	require.NoError(err)
	updateConfigReq := domain.CreateUpdateConfigRequest{
		Id:       activeConfig.Id,
		Name:     activeConfig.Name,
		ModuleId: activeConfig.ModuleId,
		Version:  activeConfig.Version,
		Data:     []byte(`{"someString": "value"}`),
	}
	err = apiCli.Invoke("config/config/create_update_config").
		JsonRequestBody(updateConfigReq).
		Do(context.Background())
	require.NoError(err)

	cancelClient2()
	err = clientA2.Close()
	require.NoError(err)
	time.Sleep(2 * time.Second)

	require.Len(eventHandler.ReceivedConfigs(), 2)
	require.EqualValues([]byte("{}"), eventHandler.ReceivedConfigs()[0])

	require.Len(eventHandler.ReceivedHosts(), 2)
	require.ElementsMatch([]string{"10.2.9.1:9999", "10.2.9.2:9999"}, eventHandler.ReceivedHosts()[0])
	require.EqualValues([]string{"10.2.9.1:9999"}, eventHandler.ReceivedHosts()[1])

	require.Len(eventHandler.ReceivedRoutes(), 3)

	statusResponse := make([]domain.ModuleInfo, 0)
	err = apiCli.Invoke("config/module/get_modules_info").
		JsonResponseBody(&statusResponse).
		Do(context.Background())
	require.NoError(err)
	require.Len(statusResponse, 3)
	require.EqualValues("A", statusResponse[0].Name)
	require.Len(statusResponse[0].Status, 1)
	require.EqualValues("B", statusResponse[1].Name)
	require.Len(statusResponse[1].Status, 1)
	require.NotEmpty(statusResponse[1].ConfigSchema)
}

func dataDir(t *testing.T) string {
	t.Helper()

	data := make([]byte, 6)
	_, _ = rand.Read(data)
	dir := hex.EncodeToString(data)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}

type testModuleRemoteConfig struct {
	SomeString string `validate:"required"`
}

func newClusterClient(
	moduleName string,
	host string,
	logger log.Logger,
) *cluster.Client {
	schema := schema.NewGenerator().Generate(testModuleRemoteConfig{})
	schemaData, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	return cluster.NewClient(cluster.ModuleInfo{
		ModuleName:    moduleName,
		ModuleVersion: "1.0.0",
		LibVersion:    "1.0.0",
		GrpcOuterAddress: cluster.AddressConfiguration{
			IP:   host,
			Port: "9999",
		},
		Endpoints: nil,
	}, cluster.ConfigData{
		Version: "1.0.0",
		Schema:  schemaData,
		Config:  []byte(`{}`),
	}, []string{"127.0.0.1:9001"}, logger)
}
