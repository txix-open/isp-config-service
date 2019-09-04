package cluster

import (
	"github.com/cenkalti/backoff"
	gosocketio "github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/logger"
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

func (c *SocketLeaderClient) Dial(timeout time.Duration) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	return backoff.Retry(c.c.Dial, backOff)
}

func (c *SocketLeaderClient) Close() {
	c.c.Close()
}

func NewSocketLeaderClient(address string) *SocketLeaderClient {
	socketIoAddress := getSocketIoUrl(address)
	client := gosocketio.NewClientBuilder().
		EnableReconnection().
		ReconnectionTimeout(1 * time.Second).
		OnReconnectionError(func(err error) {
			logger.Warnf("socket.io leader(%s) client reconnection err: %v", socketIoAddress, err)
		}).
		BuildToConnect(socketIoAddress)
	return &SocketLeaderClient{
		c: client,
	}
}

func getSocketIoUrl(address string) string {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err) //must never occured
	}

	cfg := config.Get().(*conf.Configuration)
	port, err := strconv.Atoi(cfg.WS.Rest.Port)
	if err != nil {
		panic(err)
	}
	return gosocketio.GetUrl(addr.IP.String(), port, false, map[string]string{
		ClusterParam: "true",
	})

}
