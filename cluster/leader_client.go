package cluster

import (
	"github.com/cenkalti/backoff"
	gosocketio "github.com/integration-system/golang-socketio"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/logger"
	"isp-config-service/conf"
	"net"
	"time"
)

type SocketLeaderClient struct {
	c *gosocketio.Client
}

func (c *SocketLeaderClient) Send(data []byte, timeout time.Duration) (string, error) {
	return c.c.Ack(applyCommandEvent, data, timeout)
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
	socketIoAddress := getSocketIoAddress(address)
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

func getSocketIoAddress(address string) string {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err) //must never occured
	}

	cfg := config.Get().(*conf.Configuration)
	return net.JoinHostPort(addr.IP.String(), cfg.WS.Rest.Port)
}
