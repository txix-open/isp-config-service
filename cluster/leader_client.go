package cluster

import (
	"errors"
	"github.com/cenkalti/backoff"
	gosocketio "github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/conf"
	"net"
	"strconv"
	"time"
)

type SocketLeaderClient struct {
	c *gosocketio.Client
}

func (c *SocketLeaderClient) Send(data []byte, timeout time.Duration) (string, error) {
	return c.c.Ack(ApplyCommandEvent, data, timeout)
}

func (c *SocketLeaderClient) SendDeclaration(backend structure.BackendDeclaration, timeout time.Duration) (string, error) {
	data, err := json.Marshal(backend)
	if err != nil {
		return "", err
	}
	if c.c.IsAlive() {
		return c.c.Ack(utils.ModuleReady, data, timeout)
	}
	return "", errors.New("SendDeclaration connection is not alive")
}

func (c *SocketLeaderClient) Dial(timeout time.Duration) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	return backoff.Retry(c.c.Dial, backOff)
}

func (c *SocketLeaderClient) Close() {
	c.c.Close()
}

func NewSocketLeaderClient(address string, leaderDisconnectionCallback func()) *SocketLeaderClient {
	socketIoAddress := getSocketIoUrl(address)
	client := gosocketio.NewClientBuilder().
		EnableReconnection().
		ReconnectionTimeout(1 * time.Second).
		OnReconnectionError(func(err error) {
			log.Warnf(codes.LeaderClientConnectionError, "leader client reconnection err: %v", err)
		}).
		BuildToConnect(socketIoAddress)
	err := client.On(gosocketio.OnDisconnection, func(channel *gosocketio.Channel) {
		log.WithMetadata(map[string]interface{}{
			"leaderIp": channel.Ip(),
		}).Warn(codes.LeaderClientDisconnected, "leader client disconnected")
		leaderDisconnectionCallback()
	})
	if err != nil {
		panic(err) ////must never occurred, and will removed in future
	}

	return &SocketLeaderClient{
		c: client,
	}
}

func getSocketIoUrl(address string) string {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err) //must never occurred
	}

	cfg := config.Get().(*conf.Configuration)
	port, err := strconv.Atoi(cfg.WS.Rest.Port)
	if err != nil {
		panic(err)
	}
	// TODO логика для определения порта пира, т.к всё тестируется на одной машине
	//peerNumber := addr.Port - 9000
	//switch peerNumber {
	//case 2:
	//	port = 9011
	//case 3:
	//	port = 9021
	//case 4:
	//	port = 9031
	//}
	//
	return gosocketio.GetUrl(addr.IP.String(), port, false, map[string]string{
		ClusterParam:  "true",
		"module_name": cfg.ModuleName,
		// TODO вынести ключи в константы в isp-lib и выпилить instance_uuid
		"instance_uuid": "9d89354b-c728-4b48-b002-a7d3b229f151",
	})

}
