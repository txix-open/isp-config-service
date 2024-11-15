package tests_test

import (
	"context"
	"sync"

	"github.com/txix-open/isp-kit/cluster"
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

func (c *clusterEventHandler) ReceivedConfigs() [][]byte {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.receivedConfigs
}

func (c *clusterEventHandler) ReceivedRoutes() []cluster.RoutingConfig {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.receivedRoutes
}

func (c *clusterEventHandler) ReceivedHosts() [][]string {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.receivedHosts
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
