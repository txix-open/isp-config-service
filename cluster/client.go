package cluster

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/integration-system/isp-lib/v2/atomic"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"isp-config-service/codes"
	"isp-config-service/entity"
	"isp-config-service/raft"
	"isp-config-service/store/state"
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
	isLeader           *atomic.AtomicBool
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
	return client.isLeader.Get()
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
			return nil, fmt.Errorf("apply to raft as a leader: %v", err)
		}
		logResponse := apply.(ApplyLogResponse)
		return &logResponse, nil
	}

	if client.leaderClient == nil {
		return nil, ErrLeaderClientNotInitialized
	}
	response, err := client.leaderClient.Ack(command, defaultApplyTimeout)
	if err != nil {
		return nil, fmt.Errorf("send to leader to apply: %v", err)
	}

	var logResponse ApplyLogResponse
	err = json.Unmarshal(response, &logResponse)
	if err != nil {
		return nil, err
	}

	return &logResponse, nil
}

func (client *Client) SyncApplyOnLeader(command []byte) (*ApplyLogResponse, error) {
	client.leaderMu.RLock()
	defer client.leaderMu.RUnlock()

	if !client.leaderState.isLeader {
		return nil, ErrNotLeader
	}
	apply, err := client.r.SyncApply(command)
	if err != nil {
		return nil, fmt.Errorf("apply to raft as a leader: %v", err)
	}
	logResponse := apply.(ApplyLogResponse)
	return &logResponse, nil
}

func (client *Client) listenLeader() {
	for n := range client.r.LeaderCh() {
		client.leaderMu.Lock()
		log.WithMetadata(map[string]interface{}{
			"leader_elected": n.LeaderElected,
			"current_leader": n.CurrentLeaderAddress,
			"is_leader":      n.IsLeader,
		}).Info(codes.LeaderStateChanged, "leader state changed")

		if client.leaderClient != nil {
			client.leaderClient.Close()
			client.leaderClient = nil
		}
		if n.LeaderElected && !n.IsLeader {
			leaderAddr := n.CurrentLeaderAddress
			leaderClient := NewSocketLeaderClient(leaderAddr, func() {
				client.onClientDisconnect(leaderAddr)
			})
			if err := leaderClient.Dial(leaderConnectionTimeout); err != nil {
				log.Fatalf(codes.LeaderClientConnectionError, "could not connect to leader: %v", err)
			}
			client.leaderClient = leaderClient

			log.Info(codes.SendDeclarationToLeader, "sending declaration to leader through websocket")
			go client.declareMyselfToLeader(leaderClient)
		} else if n.LeaderElected && n.IsLeader {
			go client.declareMyselfToCluster()
		}
		client.leaderState = leaderState{
			leaderElected: n.LeaderElected,
			isLeader:      n.IsLeader,
			leaderAddr:    n.CurrentLeaderAddress,
		}
		client.isLeader.Set(n.IsLeader)

		client.leaderMu.Unlock()
	}
}
func (client *Client) declareMyselfToLeader(leaderClient *SocketLeaderClient) {
	response, err := leaderClient.SendDeclaration(client.declaration, defaultApplyTimeout)
	if err != nil {
		log.Warnf(codes.SendDeclarationToLeader, "send declaration to leader. err: %v", err)
	} else if response != utils.WsOkResponse {
		log.Warnf(codes.SendDeclarationToLeader, "send declaration to leader. error response: '%s'", response)
	} else {
		log.Info(codes.SendDeclarationToLeader, "successfully sent my declaration to leader")
	}
}

// used when a leader need to declare himself to a cluster
func (client *Client) declareMyselfToCluster() {
	now := state.GenerateDate()
	module := entity.Module{
		Id:              state.GenerateId(),
		Name:            client.declaration.ModuleName,
		CreatedAt:       now,
		LastConnectedAt: now,
	}
	command := PrepareModuleConnectedCommand(module)
	syncApplyCommand(client, command)

	declarationCommand := PrepareUpdateBackendDeclarationCommand(client.declaration)
	syncApplyCommand(client, declarationCommand)
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
		isLeader:           atomic.NewAtomicBool(false),
	}
	go client.listenLeader()

	return client
}

func syncApplyCommand(clusterClient *Client, command []byte) {
	applyLogResponse, err := clusterClient.SyncApply(command)
	if err != nil {
		commandName := ParseCommand(command).String()
		log.WithMetadata(map[string]interface{}{
			"command":     string(command),
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "announce myself. apply command: %v", err)
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		commandName := ParseCommand(command).String()
		log.WithMetadata(map[string]interface{}{
			"result":      string(applyLogResponse.Result),
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "announce myself. apply command")
	}
}
