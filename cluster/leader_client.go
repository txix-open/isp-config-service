package cluster

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	etp "github.com/integration-system/isp-etp-go/client"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/conf"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	LeaderClientReconnectionTimeout = 500 * time.Millisecond
)

type SocketLeaderClient struct {
	client    etp.Client
	url       string
	globalCtx context.Context
	cancel    context.CancelFunc
}

func (c *SocketLeaderClient) Ack(data []byte, timeout time.Duration) ([]byte, error) {
	ctx, _ := context.WithTimeout(c.globalCtx, timeout)
	response, err := c.client.EmitWithAck(ctx, ApplyCommandEvent, data)
	return response, err
}

func (c *SocketLeaderClient) SendDeclaration(backend structure.BackendDeclaration, timeout time.Duration) (string, error) {
	data, err := json.Marshal(backend)
	if err != nil {
		return "", err
	}
	ctx, _ := context.WithTimeout(c.globalCtx, timeout)
	response, err := c.client.EmitWithAck(ctx, utils.ModuleReady, data)
	return string(response), err
}

func (c *SocketLeaderClient) Dial(timeout time.Duration) error {
	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = timeout
	dial := func() error {
		return c.client.Dial(context.Background(), c.url)
	}
	return backoff.Retry(dial, backOff)
}

func (c *SocketLeaderClient) Close() {
	c.cancel()
	err := c.client.Close()
	if err != nil {
		log.Warnf(codes.LeaderClientConnectionError, "leader client close err: %v", err)
	}
	log.Debug(0, "leader client connection closed")
}

func NewSocketLeaderClient(address string, leaderDisconnectionCallback func()) *SocketLeaderClient {
	etpConfig := etp.Config{
		HttpClient: http.DefaultClient,
	}
	client := etp.NewClient(etpConfig)
	leaderClient := &SocketLeaderClient{
		client: client,
		url:    getUrl(address),
	}
	ctx, cancel := context.WithCancel(context.Background())
	leaderClient.globalCtx = ctx
	leaderClient.cancel = cancel

	leaderClient.client.OnDisconnect(func(err error) {
		log.WithMetadata(map[string]interface{}{
			"leaderAddr": address,
		}).Warn(codes.LeaderClientDisconnected, "leader client disconnected")
		leaderDisconnectionCallback()
		go func() {
			for {
				err := leaderClient.client.Dial(leaderClient.globalCtx, leaderClient.url)
				if err == nil {
					log.Warnf(codes.LeaderClientConnectionError, "leader client reconnected")
					return
				} else if errors.Is(err, context.Canceled) {
					log.Warnf(codes.LeaderClientConnectionError, "leader client reconnection canceled")
					return
				} else {
					log.Warnf(codes.LeaderClientConnectionError, "leader client reconnection err: %v", err)
				}
				time.Sleep(LeaderClientReconnectionTimeout)
			}
		}()
	})

	leaderClient.client.OnError(func(err error) {
		log.Warnf(codes.LeaderClientConnectionError, "leader client on error: %v", err)
	})
	leaderClient.client.OnConnect(func() {
		log.Warnf(codes.LeaderClientConnectionError, "leader client connected")
	})
	return leaderClient
}

func getUrl(address string) string {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		panic(err) // must never occurred
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
	params := url.Values{}
	params.Add(ClusterParam, "true")
	// TODO вынести ключи в константы в isp-lib и выпилить instance_uuid
	params.Add("module_name", cfg.ModuleName)
	params.Add("instance_uuid", "9d89354b-c728-4b48-b002-a7d3b229f151")
	return fmt.Sprintf("ws://%s:%d/isp-etp/?%s", addr.IP.String(), port, params.Encode())
}
