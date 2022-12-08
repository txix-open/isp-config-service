package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib-test/docker"
	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib-test/utils/postgres"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/bootstrap"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-config-service/conf"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/store/state"
)

const (
	configServiceHttpPort = "9001"
	configServiceGrpcPort = "9002"
	configServiceSchema   = "config_service"

	configsNumber   = 3
	maxAwaitingTime = 25 * time.Second
	attemptTimeout  = 300 * time.Millisecond

	deleteCommonConfigsCommand  = "config/common_config/delete_config"
	getRoutesCommand            = "config/routing/get_routes"
	getModulesInfoCommand       = "config/module/get_modules_info"
	createUpdateConfigCommand   = "config/config/create_update_config"
	getAllConfigVersionsCommand = "config/config/get_all_version"
)

var (
	pgCfg structure.DBConfiguration

	configsHttpAddrs = make([]structure.AddressConfiguration, configsNumber)
	configsGrpcAddrs = make([]structure.AddressConfiguration, configsNumber)
	configsCtxs      = make([]*docker.ContainerContext, configsNumber)
)

const mockModuleName = "mockModule"

type (
	mockRemoteConfig struct {
		Value1 string `json:"value1"`
		Value2 int    `json:"value2"`
		Value3 Value3 `json:"value3"`
	}
	Value3 struct {
		Value31 string `json:"value31"`
	}
)

func testMain(m *testing.M) {
	cfg := ctx.BaseTestConfiguration{}
	test, err := ctx.NewIntegrationTest(m, &cfg, setup)
	if err != nil {
		panic(err)
	}
	test.PrepareAndRun()
}

func setup(testCtx *ctx.TestContext, runTest func() int) int {
	cfg := testCtx.BaseConfiguration()

	cli, err := docker.NewClient()
	if err != nil {
		panic(err)
	}
	defer cli.Close()
	env := docker.NewTestEnvironment(testCtx, cli)
	defer env.Cleanup()

	_, pgCfg = env.RunPGContainer()
	_, err = postgres.Wait(pgCfg, 10*time.Second)
	if err != nil {
		panic(err)
	}
	pgCfg.Schema = configServiceSchema

	configs := make([]conf.Configuration, configsNumber)
	peersAddrs := make([]string, configsNumber)

	for i := 0; i < configsNumber; i++ {
		httpAddr := getConfigServiceAddress(i, configServiceHttpPort)
		grpcAddr := getConfigServiceAddress(i, configServiceGrpcPort)
		cfg := conf.Configuration{
			Database:         pgCfg,
			ModuleName:       moduleName,
			GrpcOuterAddress: grpcAddr,
			WS: conf.WebService{
				Rest: httpAddr,
				Grpc: grpcAddr,
			},
			Cluster: conf.ClusterConfiguration{
				InMemory:              false,
				DataDir:               "./data",
				Peers:                 nil,
				OuterAddress:          httpAddr.GetAddress(),
				ConnectTimeoutSeconds: 10,
				BootstrapCluster:      false,
			},
		}
		configsGrpcAddrs[i] = grpcAddr
		configsHttpAddrs[i] = httpAddr
		configs[i] = cfg
		peersAddrs[i] = httpAddr.GetAddress()
	}

	configs[0].Cluster.BootstrapCluster = true

	for i := 0; i < configsNumber; i++ {
		c := configs[i]
		peersAddrsEnv := generatePeersConfigEnv(peersAddrs)
		peersAddrsEnv["LOG_LEVEL"] = "debug"
		ctx := env.RunAppContainer(
			cfg.Images.Module, c, nil,
			docker.WithName(c.GrpcOuterAddress.IP),
			docker.WithLogger(NewWriteLogger(strconv.Itoa(i)+"_config:", ioutil.Discard, "DeleteCommonConfigsCommand")),
			// docker.WithLogger(NewWriteLogger(strconv.Itoa(i)+"_config:", ioutil.Discard, "Apply 10 command")),
			// docker.PullImage(cfg.Registry.Username, cfg.Registry.Password),
			docker.WithEnv(peersAddrsEnv),
		)
		grpcAddr := configsGrpcAddrs[i]
		grpcAddr.IP = ctx.GetIPAddress()
		configsGrpcAddrs[i] = grpcAddr

		httpAddr := configsHttpAddrs[i]
		httpAddr.IP = ctx.GetIPAddress()
		configsHttpAddrs[i] = httpAddr

		configsCtxs[i] = ctx
	}
	//nolint
	// time.Sleep(3 * time.Second)
	return runTest()
}

func TestClusterElection(t *testing.T) {
	a := assert.New(t)
	ready := testClusterReady(a, -1)
	if !ready {
		return
	}

	for i := 0; i < configsNumber; i++ {
		fmt.Print("\n\n\n")
		log.Printf("stopping %d container\n", i)
		a.NoError(configsCtxs[i].StopContainer(20 * time.Second))

		time.Sleep(5 * time.Second)
		fmt.Print("\n\n\n")
		log.Printf("checking cluster except %d\n", i)
		ready = testClusterReady(a, i)
		if !ready {
			return
		}

		fmt.Print("\n\n\n")
		log.Printf("starting %d container\n", i)
		a.NoError(configsCtxs[i].StartContainer())

		time.Sleep(4 * time.Second)
		log.Printf("checking cluster, iteration %d\n", i+1)
		ready = testClusterReady(a, -1)
		if !ready {
			return
		}
	}
}

//nolint:funlen
func TestUpdateConfig(t *testing.T) {
	a := assert.New(t)
	ready := testClusterReady(a, -1)
	if !ready {
		return
	}

	httpAddr := configsHttpAddrs[configsNumber-1]
	grpcAddr := configsGrpcAddrs[configsNumber-1]

	remoteConfig := mockRemoteConfig{
		Value1: "sad",
		Value2: 6,
		Value3: Value3{
			Value31: "32",
		},
	}
	tempDir := t.TempDir()
	remoteCfgPath := filepath.Join(tempDir, "remote_config.json")
	remoteCfgJson, err := json.Marshal(remoteConfig)
	a.NoError(err)
	a.NoError(os.WriteFile(remoteCfgPath, remoteCfgJson, 0666))
	go newMockModule(a, httpAddr, remoteCfgPath)
	time.Sleep(5 * time.Second) // TODO

	client := newGrpcClient(a, grpcAddr)
	if client == nil {
		a.FailNow("unable to connect to config")
	}

	getModuleConfig := func() entity.Config {
		modulesInfo := getModulesInfo(a, client)
		a.Len(modulesInfo, 2)
		mockModule := modulesInfo[1]
		a.Equal(mockModuleName, mockModule.Name)
		if !a.Len(mockModule.Configs, 1) {
			t.FailNow()
		}
		return mockModule.Configs[0].Config
	}
	initialConfig := getModuleConfig()

	initialConfigVersions := getAllConfigVersions(a, client, initialConfig.Id)
	a.Len(initialConfigVersions, 0)
	a.EqualValues(1, initialConfig.Version)
	a.EqualValues(remoteConfig.Value2, initialConfig.Data["value2"])

	remCfg := map[string]interface{}{
		"value1": "test1",
		"value2": 0,
		"value3": map[string]interface{}{
			"value31": "something",
		},
	}
	configForUpdate := entity.Config{
		Id:          initialConfig.Id,
		Name:        initialConfig.Name,
		ModuleId:    initialConfig.ModuleId,
		Description: initialConfig.Description,
		Data:        remCfg,
	}
	const updatesCount = state.DefaultVersionCount + 1
	for i := 0; i < updatesCount; i++ {
		remCfg["value2"] = i
		updatedConfig := updateConfig(a, client, configForUpdate)
		a.EqualValues(i, updatedConfig.Data["value2"])
	}
	// wait until raft is synchronized to non-leader nodes
	time.Sleep(3 * time.Second) // TODO remove?

	const lastVersion = updatesCount + 1
	const lastDataValue = updatesCount - 1
	finalConfig := getModuleConfig()
	a.EqualValues(lastVersion, finalConfig.Version)
	a.EqualValues(lastDataValue, finalConfig.Data["value2"])

	configVersions := getAllConfigVersions(a, client, initialConfig.Id)
	a.Len(configVersions, state.DefaultVersionCount)
	for i, cfgVer := range configVersions {
		// -1 because last config is not listed in versions
		version := lastVersion - i - 1
		dataValue := lastDataValue - i - 1
		a.EqualValues(dataValue, cfgVer.Data["value2"])
		a.EqualValues(version, cfgVer.ConfigVersion)
	}
}

func getConfigServiceAddress(num int, port string) structure.AddressConfiguration {
	lastIP := pgCfg.Address
	ip := net.ParseIP(lastIP).To4()
	ip[3] = byte(int(ip[3]) + num + 1)
	return structure.AddressConfiguration{
		//IP:   fmt.Sprintf("%s-%d", "isp-config-service", num),
		IP:   ip.String(),
		Port: port,
	}
}

func testClusterReady(a *assert.Assertions, except int) bool {
	clients := make([]*backend.RxGrpcClient, 0, configsNumber)
	defer func() {
		for _, client := range clients {
			_ = client.Close()
		}
	}()

	for j := 0; j < configsNumber; j++ {
		if j == except {
			continue
		}
		configAddr := configsGrpcAddrs[j]
		client := newGrpcClient(a, configAddr)
		if client == nil {
			a.Failf("", "unable to connect to %d config", j)
			return false
		}
		clients = append(clients, client)
		ready := testRaftReady(a, client)
		if !ready {
			return ready
		}
	}

	prevRoutes := getRoutes(a, clients[0])
	routesLen := len(clients)
	for i := 1; i < len(clients); i++ {
		routes := getRoutes(a, clients[i])
		t := a.Len(routes, routesLen, "invalid length")
		if !t {
			return t
		}
		match := a.ElementsMatch(prevRoutes, routes, "different backend declarations")
		if !match {
			return match
		}
	}

	return true
}

func newGrpcClient(a *assert.Assertions, configAddr structure.AddressConfiguration) *backend.RxGrpcClient {
	start := time.Now()
	client := backend.NewRxGrpcClient(backend.WithDialOptions(
		grpc.WithInsecure(),
	))
	client.ReceiveAddressList([]structure.AddressConfiguration{configAddr})

	_, err := await(func() (interface{}, error) {
		err := client.Invoke(getRoutesCommand, -1, nil, nil)
		code := status.Code(err)
		if code != codes.Unknown && code != codes.Unavailable {
			return nil, nil
		}
		//nolint
		// log.Println("connect to grpc err:", err)
		return nil, err
	}, maxAwaitingTime, attemptTimeout)

	if !a.NoError(err) {
		return nil
	}
	log.Println("waiting until grpc ready:", time.Since(start).Round(time.Second))
	return client
}

func testRaftReady(a *assert.Assertions, client *backend.RxGrpcClient) bool {
	req := new(domain.ConfigIdRequest)
	req.Id = "33"
	response := new(structure.GrpcError)
	start := time.Now()
	f := func() (interface{}, error) {
		err := client.Invoke(
			deleteCommonConfigsCommand,
			-1,
			req,
			response,
		)
		//nolint
		// log.Println("send grpc request. response: ", response, "err: ", err)
		return nil, err
	}
	_, err := await(f, maxAwaitingTime, attemptTimeout)
	log.Println("waiting until raft ready:", time.Since(start).Round(time.Second))
	return a.NoError(err)
}

func getRoutes(a *assert.Assertions, client *backend.RxGrpcClient) []structure.BackendDeclaration {
	f := func() (interface{}, error) {
		var response []structure.BackendDeclaration
		err := client.Invoke(
			getRoutesCommand,
			-1,
			nil,
			&response,
		)
		if err != nil {
			return nil, err
		} else if len(response) == 0 {
			return nil, errors.New("zero routes")
		}
		//nolint
		// log.Println("routes:", response)
		return response, nil
	}

	response, err := await(f, time.Second, 50*time.Millisecond)
	a.NoError(err)
	declarations, _ := response.([]structure.BackendDeclaration)
	return declarations
}

func getModulesInfo(a *assert.Assertions, client *backend.RxGrpcClient) []domain.ModuleInfo {
	var response []domain.ModuleInfo
	err := client.Invoke(
		getModulesInfoCommand,
		-1,
		nil,
		&response,
	)

	a.NoError(err)
	return response
}

func updateConfig(a *assert.Assertions, client *backend.RxGrpcClient, config entity.Config) domain.ConfigModuleInfo {
	var response domain.ConfigModuleInfo
	err := client.Invoke(
		createUpdateConfigCommand,
		-1,
		domain.CreateUpdateConfigRequest{Config: config},
		&response,
	)

	a.NoError(err)
	return response
}

func getAllConfigVersions(a *assert.Assertions, client *backend.RxGrpcClient, configId string) []entity.VersionConfig {
	var response []entity.VersionConfig
	err := client.Invoke(
		getAllConfigVersionsCommand,
		-1,
		domain.ConfigIdRequest{Id: configId},
		&response,
	)

	a.NoError(err)
	return response
}

func await(dialer func() (interface{}, error), timeout, attemptTimeout time.Duration) (interface{}, error) {
	retryer := utils.NewRetryer(dialer, timeout)
	retryer.AttemptTimeout = attemptTimeout
	return retryer.Do()
}

func generatePeersConfigEnv(peers []string) map[string]string {
	key := "LC_ISP_CLUSTER.PEERS"
	val := strings.Join(peers, ",")
	return map[string]string{
		key: val,
	}
}

type writeLogger struct {
	prefix string
	w      io.Writer
	filter string
}

func (l *writeLogger) Write(p []byte) (int, error) {
	lines := bytes.SplitAfter(p, []byte("\n"))
	for _, line := range lines {
		s := string(line)
		if l.filter != "" && strings.Contains(s, l.filter) || strings.TrimSpace(s) == "" {
			continue
		}
		_, err := l.w.Write(p)
		if err != nil {
			fmt.Printf("%s %s: %v", l.prefix, s, err)
		} else {
			fmt.Printf("%s %s", l.prefix, s)
		}
	}
	return len(p), nil
}

// NewWriteLogger returns a writer that behaves like w except
// that it logs (using fmt.Printf) each write to standard error,
// printing the prefix and the string data written.
func NewWriteLogger(prefix string, w io.Writer, filter string) io.Writer {
	return &writeLogger{prefix, w, filter}
}

func newMockModule(a *assert.Assertions, configAddr structure.AddressConfiguration, remoteCfgPath string) {
	makeDeclaration := func(localConfig interface{}) bootstrap.ModuleInfo {
		return bootstrap.ModuleInfo{
			ModuleName:    mockModuleName,
			ModuleVersion: "dev",
			GrpcOuterAddress: structure.AddressConfiguration{
				IP:   "128.3.3.3",
				Port: "9999",
			},
			Endpoints: nil,
		}
	}

	socketConfiguration := func(cfg interface{}) structure.SocketConfiguration {
		return structure.SocketConfiguration{
			Host:   configAddr.IP,
			Port:   configAddr.Port,
			Secure: false,
			UrlParams: map[string]string{
				"module_name": mockModuleName,
			},
		}
	}

	cfg := bootstrap.
		ServiceBootstrap(&conf.Configuration{}, &mockRemoteConfig{}).
		DefaultRemoteConfigPath(remoteCfgPath).
		SocketConfiguration(socketConfiguration).
		OnSocketErrorReceive(func(errorMessage map[string]interface{}) {
			a.Fail("OnSocketErrorReceive", errorMessage)
		}).
		OnConfigErrorReceive(func(errorMessage string) {
			fmt.Printf("OnConfigErrorReceive: %s\n", errorMessage)
		}).
		DeclareMe(makeDeclaration).
		OnShutdown(func(ctx context.Context, sig os.Signal) {
			a.Fail("OnShutdown")
		}).
		OnRemoteConfigReceive(func(remoteConfig, _ *mockRemoteConfig) {
			// TODO: compare configs?
			fmt.Println("OnRemoteConfigReceive", remoteConfig)
		}).OnModuleReady(func() {
		fmt.Println("OnModuleReady")
	})

	cfg.Run()
}
