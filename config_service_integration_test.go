package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/integration-system/isp-lib-test/ctx"
	"github.com/integration-system/isp-lib-test/docker"
	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib-test/utils/postgres"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-config-service/conf"
	"isp-config-service/domain"
)

const (
	configServiceHttpPort = "9001"
	configServiceGrpcPort = "9002"
	configServiceSchema   = "config_service"

	configsNumber   = 3
	maxAwaitingTime = 25 * time.Second
	attemptTimeout  = 300 * time.Millisecond

	deleteCommonConfigsCommand = "config/common_config/delete_config"
	getRoutesCommand           = "config/routing/get_routes"
)

var (
	pgCfg structure.DBConfiguration

	configsHttpAddrs = make([]structure.AddressConfiguration, configsNumber)
	configsGrpcAddrs = make([]structure.AddressConfiguration, configsNumber)
	configsCtxs      = make([]*docker.ContainerContext, configsNumber)
)

func TestMain(m *testing.M) {
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
			//docker.WithLogger(NewWriteLogger(strconv.Itoa(i)+"_config:", ioutil.Discard, "Apply 10 command")),
			//docker.PullImage(cfg.Registry.Username, cfg.Registry.Password),
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

	//time.Sleep(3 * time.Second)
	return runTest()
}

func TestClusterElection(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	a := assert.New(t)
	var ready bool

	ready = testClusterReady(a, -1)
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

		time.Sleep(3 * time.Second)
		log.Printf("checking cluster, iteration %d\n", i+1)
		ready = testClusterReady(a, -1)
		if !ready {
			return
		}
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
	var clients []*backend.RxGrpcClient
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
		client := newGrpcClient(configAddr, a)
		if client == nil {
			a.Fail(fmt.Sprintf("unable to connect to %d config", j))
			return false
		}
		clients = append(clients, client)
		ready := testRaftReady(client, a)
		if !ready {
			return ready
		}
	}

	prevRoutes := getRoutes(clients[0], a)
	routesLen := len(clients)
	for i := 1; i < len(clients); i++ {
		routes := getRoutes(clients[i], a)
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

func newGrpcClient(configAddr structure.AddressConfiguration, a *assert.Assertions) *backend.RxGrpcClient {
	start := time.Now()
	client := backend.NewRxGrpcClient(backend.WithDialOptions(
		grpc.WithInsecure(),
	))
	client.ReceiveAddressList([]structure.AddressConfiguration{configAddr})

	_, err := await(func() (interface{}, error) {
		err := client.Invoke(getRoutesCommand, -1, nil, nil)
		code := status.Code(err)
		if code != codes.Unknown {
			return nil, nil
		}

		//log.Println("connect to grpc err:", err)
		return nil, err
	}, maxAwaitingTime, attemptTimeout)

	if !a.NoError(err) {
		return nil
	}
	log.Println("waiting until grpc ready:", time.Since(start).Round(time.Second))
	return client
}

func testRaftReady(client *backend.RxGrpcClient, a *assert.Assertions) bool {
	var err error
	req := new(domain.ConfigIdRequest)
	req.Id = "33"
	response := new(structure.GrpcError)
	start := time.Now()
	f := func() (interface{}, error) {
		err = client.Invoke(
			deleteCommonConfigsCommand,
			-1,
			req,
			response,
		)
		//log.Println("send grpc request. response: ", response, "err: ", err)
		return nil, err
	}
	_, _ = await(f, maxAwaitingTime, attemptTimeout)
	log.Println("waiting until raft ready:", time.Since(start).Round(time.Second))
	return a.NoError(err)
}

func getRoutes(client *backend.RxGrpcClient, a *assert.Assertions) []structure.BackendDeclaration {
	var response []structure.BackendDeclaration
	var err error
	f := func() (interface{}, error) {
		err = client.Invoke(
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

		//log.Println("routes:", response)
		return nil, nil
	}

	_, _ = await(f, time.Second, 50*time.Millisecond)
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
