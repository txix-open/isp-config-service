package cluster

import (
	"errors"
	"github.com/integration-system/isp-lib/logger"
	"isp-config-service/raft"
	"sync"
	"time"
)

var (
	ErrNoLeader                   = errors.New("no leader in cluster")
	ErrLeaderClientNotInitialized = errors.New("leader client not initialized")
	ErrNotLeader                  = errors.New("node is not a leader")
)

const (
	leaderConnectionTimeout = 3 * time.Second
	defaultApplyTimeout     = 3 * time.Second
)

type ClusterClient struct {
	r *raft.Raft

	leaderMu     sync.RWMutex
	leaderState  leaderState
	leaderClient *SocketLeaderClient
}

func (client *ClusterClient) Shutdown() error {
	client.leaderMu.Lock()
	defer client.leaderMu.Unlock()

	if client.leaderClient != nil {
		client.leaderClient.Close()
		client.leaderClient = nil
	}

	return client.r.GracefulShutdown()
}

func (client *ClusterClient) SyncApply(command []byte) (interface{}, error) {
	client.leaderMu.RLock()
	defer client.leaderMu.RUnlock()

	if !client.leaderState.leaderElected {
		return nil, ErrNoLeader
	}

	if client.leaderState.isLeader {
		return client.r.SyncApply(command)
	} else {
		if client.leaderClient == nil {
			return nil, ErrLeaderClientNotInitialized
		}
		return client.leaderClient.Send(command, defaultApplyTimeout)
	}
}

func (client *ClusterClient) SyncApplyOnLeader(command []byte) (interface{}, error) {
	client.leaderMu.RLock()
	defer client.leaderMu.RUnlock()

	if !client.leaderState.isLeader {
		return nil, ErrNotLeader
	}

	return client.r.SyncApply(command)
}

func (client *ClusterClient) listenLeader() {
	for n := range client.r.LeaderCh() {
		logger.Debug("ChangeLeaderNotification:", n)

		client.leaderMu.Lock()
		if client.leaderState.leaderAddr != n.CurrentLeaderAddress {
			if client.leaderClient != nil {
				client.leaderClient.Close()
				client.leaderClient = nil
			}
		}
		if n.LeaderElected && !n.IsLeader {
			leaderClient := NewSocketLeaderClient(n.CurrentLeaderAddress)
			if err := leaderClient.Dial(leaderConnectionTimeout); err != nil {
				logger.Fatalf("could not connect to leader: %v", err)
				continue
			}
			client.leaderClient = leaderClient
		}
		client.leaderState = leaderState{
			leaderElected: n.LeaderElected,
			isLeader:      n.IsLeader,
			leaderAddr:    n.CurrentLeaderAddress,
		}

		client.leaderMu.Unlock()
	}
}

type leaderState struct {
	leaderElected bool
	isLeader      bool
	leaderAddr    string
}

func NewRaftClusterClient(r *raft.Raft) *ClusterClient {
	client := &ClusterClient{
		r:           r,
		leaderState: leaderState{},
	}
	go client.listenLeader()

	return client
}
