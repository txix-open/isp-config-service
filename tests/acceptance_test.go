package tests_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
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

type clusterEventHandler struct {
	receivedConfigs [][]byte
	receivedHosts   [][]string
	receivedRoutes  []cluster.RoutingConfig
	lock            sync.Locker
}

func newClusterEventHandler() *clusterEventHandler {
	return &clusterEventHandler{
		lock: &sync.Mutex{},
	}
}

func (c *clusterEventHandler) ReceiveConfig(ctx context.Context, remoteConfig []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.receivedConfigs = append(c.receivedConfigs, remoteConfig)
	return nil
}

func (c *clusterEventHandler) Upgrade(hosts []string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.receivedHosts = append(c.receivedHosts, hosts)
}

func (c *clusterEventHandler) ReceiveRoutes(ctx context.Context, routes cluster.RoutingConfig) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.receivedRoutes = append(c.receivedRoutes, routes)
	return nil
}

func TestAcceptance(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	err := os.Setenv("APP_CONFIG_PATH", "../conf/config.yml")
	require.NoError(err)
	err = os.Setenv("DefaultRemoteConfigPath", "../conf/default_remote_config.json")
	require.NoError(err)
	boot := bootstrap.New("1.0.0", conf.Remote{}, nil)
	boot.App.Logger().SetLevel(log.DebugLevel)
	boot.MigrationsDir = "../migrations"
	dataPath := dataDir(t)
	boot.App.Config().Set("rqlite.DataPath", dataPath)
	logger := boot.App.Logger()

	startup := startup.New(boot)
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
		require.NoError(err)
	}()

	clientA2 := newClusterClient("A", "10.2.9.2", logger)
	clientA2Ctx, cancelClient2 := context.WithCancel(context.Background())
	go func() {
		handler := cluster.NewEventHandler()
		err := clientA2.Run(clientA2Ctx, handler)
		require.NoError(err)
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
		require.NoError(err)
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

	require.Len(eventHandler.receivedConfigs, 2)
	require.EqualValues([]byte("{}"), eventHandler.receivedConfigs[0])

	require.Len(eventHandler.receivedHosts, 2)
	require.EqualValues([]string{"10.2.9.1:9999", "10.2.9.2:9999"}, eventHandler.receivedHosts[0])
	require.EqualValues([]string{"10.2.9.1:9999"}, eventHandler.receivedHosts[1])

	require.Len(eventHandler.receivedRoutes, 3)

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
