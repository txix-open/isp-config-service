package cluster

import (
	"errors"
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"isp-config-service/codes"
	"isp-config-service/raft"
	"strconv"
	"sync"
	"time"
)

var (
	ErrNoLeader                   = errors.New("no leader in cluster")
	ErrLeaderClientNotInitialized = errors.New("leader client not initialized")
	ErrNotLeader                  = errors.New("node is not a leader")
	json                          = jsoniter.ConfigFastest
)

const (
	leaderConnectionTimeout = 3 * time.Second
	defaultApplyTimeout     = 3 * time.Second
)

type Client struct {
	r *raft.Raft

	leaderMu           sync.RWMutex
	leaderState        leaderState
	leaderClient       *SocketLeaderClient
	declaration        structure.BackendDeclaration
	onClientDisconnect func(string)
}

func (client *Client) Shutdown() error {
	client.leaderMu.Lock()
	defer client.leaderMu.Unlock()

	if client.leaderClient != nil {
		client.leaderClient.Close()
		client.leaderClient = nil
	}

	return client.r.GracefulShutdown()
}

func (client *Client) IsLeader() bool {
	client.leaderMu.RLock()
	defer client.leaderMu.RUnlock()

	return client.leaderState.isLeader
}

func (client *Client) SyncApply(command []byte) (*ApplyLogResponse, error) {
	client.leaderMu.RLock()
	defer client.leaderMu.RUnlock()

	if !client.leaderState.leaderElected {
		return nil, ErrNoLeader
	}

	if client.leaderState.isLeader {
		apply, err := client.r.SyncApply(command)
		if err != nil {
			return nil, err
		}
		logResponse := apply.(ApplyLogResponse)
		return &logResponse, err
	} else {
		if client.leaderClient == nil {
			return nil, ErrLeaderClientNotInitialized
		}
		response, err := client.leaderClient.Send(command, defaultApplyTimeout)
		if err != nil {
			return nil, err
		}
		var logResponse ApplyLogResponse
		err = json.Unmarshal([]byte(response), &logResponse)
		if err != nil {
			return nil, err
		}
		return &logResponse, nil
	}
}

func (client *Client) SyncApplyOnLeader(command []byte) (*ApplyLogResponse, error) {
	client.leaderMu.RLock()
	defer client.leaderMu.RUnlock()

	if !client.leaderState.isLeader {
		return nil, ErrNotLeader
	}
	apply, err := client.r.SyncApply(command)
	if err != nil {
		return nil, err
	}
	logResponse := apply.(ApplyLogResponse)
	return &logResponse, err
}

func (client *Client) listenLeader() {
	for n := range client.r.LeaderCh() {
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
				log.Fatalf(codes.LeaderClientConnectionError, "could not connect to leader: %v", err)
				continue
			}
			go func(declaration structure.BackendDeclaration) {
				response, err := leaderClient.SendDeclaration(declaration, defaultApplyTimeout)
				if res, err := strconv.Unquote(response); err == nil {
					response = res
				}
				if err != nil {
					log.Warnf(codes.SendDeclarationToLeaderError, "send declaration to leader err: %v", err)
				} else if response != Ok {
					log.Warnf(codes.SendDeclarationToLeaderError, "send declaration to leader response: %s", response)
				}
			}(client.declaration)

			client.leaderClient = leaderClient
		} else if n.LeaderElected && n.IsLeader {
			go func(declaration structure.BackendDeclaration) {
				command := PrepareUpdateBackendDeclarationCommand(declaration)

				applyLogResponse, err := client.SyncApply(command)
				if err != nil {
					log.Warnf(codes.SyncApplyError, "cluster.SyncApply announce myself: %v", err)
				}
				if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
					log.WithMetadata(map[string]interface{}{
						"comment":    applyLogResponse.Result,
						"applyError": applyLogResponse.ApplyError,
					}).Warn(codes.SyncApplyError, "cluster.SyncApply announce myself")
				}
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

func NewRaftClusterClient(r *raft.Raft, declaration structure.BackendDeclaration, onLeaderDisconnect func(string)) *Client {
	client := &Client{
		r:                  r,
		declaration:        declaration,
		leaderState:        leaderState{},
		onClientDisconnect: onLeaderDisconnect,
	}
	go client.listenLeader()

	return client
}
