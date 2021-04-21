package cluster

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
	etp "github.com/integration-system/isp-etp-go/v2/client"
	"github.com/integration-system/isp-lib/v2/config"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/conf"
)

var (
	// Set from config in main
	WsConnectionReadLimit int64
)

type SocketLeaderClient struct {
	client    etp.Client
	url       string
	globalCtx context.Context
	cancel    context.CancelFunc
}

func (c *SocketLeaderClient) Ack(data []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(c.globalCtx, timeout)
	defer cancel()
	response, err := c.client.EmitWithAck(ctx, ApplyCommandEvent, data)
	return response, err
}

func (c *SocketLeaderClient) SendDeclaration(backend structure.BackendDeclaration, timeout time.Duration) (string, error) {
	data, err := json.Marshal(backend)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(c.globalCtx, timeout)
	defer cancel()
	response, err := c.client.EmitWithAck(ctx, utils.ModuleReady, data)
	return string(response), err
}

func (c *SocketLeaderClient) Dial(timeout time.Duration) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	bf := backoff.WithContext(backOff, c.globalCtx)
	dial := func() error {
		return c.client.Dial(c.globalCtx, c.url)
	}
	return backoff.Retry(dial, bf)
}

func (c *SocketLeaderClient) Close() {
	c.cancel()
	err := c.client.Close()
	if err != nil {
		log.Warnf(codes.LeaderClientConnectionError, "leader client %s close err: %v", c.url, err)
	}
	log.Debugf(0, "close previous leader ws connection %s", c.url)
}

func NewSocketLeaderClient(address string, leaderDisconnectionCallback func()) *SocketLeaderClient {
	etpConfig := etp.Config{
		HttpClient:          http.DefaultClient,
		ConnectionReadLimit: WsConnectionReadLimit,
	}
	client := etp.NewClient(etpConfig)
	leaderClient := &SocketLeaderClient{
		client: client,
		url:    getURL(address),
	}
	ctx, cancel := context.WithCancel(context.Background())
	leaderClient.globalCtx = ctx
	leaderClient.cancel = cancel

	leaderClient.client.OnDisconnect(func(err error) {
		log.WithMetadata(map[string]interface{}{
			"leaderAddr": address,
			"error":      err,
		}).Warn(codes.LeaderClientDisconnected, "disconnected ws from leader")
		leaderDisconnectionCallback()
	})

	leaderClient.client.OnError(func(err error) {
		log.WithMetadata(map[string]interface{}{
			"leaderAddr": address,
			"error":      err,
		}).Warn(codes.LeaderClientDisconnected, "leader client on error")
	})
	leaderClient.client.OnConnect(func() {
		log.WithMetadata(map[string]interface{}{
			"leaderAddr": address,
		}).Info(codes.LeaderClientConnected, "connected to new leader")
	})
	return leaderClient
}

func getURL(address string) string {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err) // must never occurred
	}
	cfg := config.Get().(*conf.Configuration)

	params := url.Values{}
	params.Add(GETParam, "true")
	// TODO выпилить instance_uuid ?
	params.Add(utils.ModuleNameGetParamKey, cfg.ModuleName)
	params.Add(utils.InstanceUuidGetParamKey, "9d89354b-c728-4b48-b002-a7d3b229f151")
	return fmt.Sprintf("ws://%s:%d/isp-etp/?%s", addr.IP.String(), addr.Port, params.Encode())
}
