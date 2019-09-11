package cluster

import (
	"errors"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/raft"
	"strconv"
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

	leaderMu           sync.RWMutex
	leaderState        leaderState
	leaderClient       *SocketLeaderClient
	declaration        structure.BackendDeclaration
	onClientDisconnect func(string)
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
			leaderClient := NewSocketLeaderClient(n.CurrentLeaderAddress, func() {
				client.onClientDisconnect(client.leaderState.leaderAddr)
			})
			if err := leaderClient.Dial(leaderConnectionTimeout); err != nil {
				logger.Fatalf("could not connect to leader: %v", err)
				continue
			}
			go func(declaration structure.BackendDeclaration) {
				response, err := leaderClient.SendDeclaration(declaration, defaultApplyTimeout)
				if res, err := strconv.Unquote(response); err == nil {
					response = res
				}
				if err != nil {
					logger.Warn("leaderClient.SendDeclaration", err)
				} else if response != "ok" {
					logger.Warn("leaderClient.SendDeclaration response", response)
				}
			}(client.declaration)

			client.leaderClient = leaderClient
		} else if n.LeaderElected && n.IsLeader {
			go func(declaration structure.BackendDeclaration) {
				command := PrepareUpdateBackendDeclarationCommand(declaration)
				i, err := client.SyncApply(command)
				logger.Debug("cluster.SyncApply announce myself:", i, err)
			}(client.declaration)
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

func NewRaftClusterClient(r *raft.Raft, declaration structure.BackendDeclaration, onLeaderDisconnect func(string)) *ClusterClient {
	client := &ClusterClient{
		r:                  r,
		declaration:        declaration,
		leaderState:        leaderState{},
		onClientDisconnect: onLeaderDisconnect,
	}
	go client.listenLeader()

	return client
}
