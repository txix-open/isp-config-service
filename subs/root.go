package subs

import (
	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"
)

var (
	json = jsoniter.ConfigFastest
)

type SocketEventHandler struct {
	server etp.Server
	store  *store.Store
}

func (h *SocketEventHandler) SubscribeAll() {
	h.server.
		OnConnect(h.handleConnect).
		OnDisconnect(h.handleDisconnect).
		OnError(h.handleError).
		OnWithAck(cluster.ApplyCommandEvent, h.applyCommandOnLeader).
		OnWithAck(utils.ModuleReady, h.handleModuleReady).
		OnWithAck(utils.ModuleSendRequirements, h.handleModuleRequirements).
		OnWithAck(utils.ModuleSendConfigSchema, h.handleConfigSchema)
}

func (h *SocketEventHandler) handleConnect(conn etp.Conn) {
	isClusterNode := IsConfigClusterNode(conn)
	moduleName, err := Parameters(conn)
	log.Debugf(0, "handleConnect from %s", moduleName) // REMOVE
	if err != nil {
		EmitConn(conn, utils.ErrorConnection, FormatErrorConnection(err))
		_ = conn.Close()
		return
	}
	holder.EtpServer.Rooms().Join(conn, service.Room.Module(moduleName))
	now := state.GenerateDate()
	module := entity.Module{
		Id:              state.GenerateId(),
		Name:            moduleName,
		CreatedAt:       now,
		LastConnectedAt: now,
	}
	command := cluster.PrepareModuleConnectedCommand(module)
	_, err = SyncApplyCommand(command)
	if err != nil && !isClusterNode {
		EmitConn(conn, utils.ErrorConnection, FormatErrorConnection(err))
		_ = conn.Close()
		return
	}

	if isClusterNode {
		return
	}

	var config map[string]interface{}
	h.store.VisitReadonlyState(func(state state.ReadonlyState) {
		config, err = service.ConfigService.GetCompiledConfig(moduleName, state)
	})
	if err != nil {
		EmitConn(conn, utils.ConfigError, []byte(err.Error()))
		return
	}
	data, err := json.Marshal(config)
	if err != nil {
		EmitConn(conn, utils.ConfigError, []byte(err.Error()))
		return
	}
	EmitConn(conn, utils.ConfigSendConfigWhenConnected, data)
}

func (h *SocketEventHandler) handleDisconnect(conn etp.Conn, _ error) {
	moduleName, _ := Parameters(conn)
	log.Debugf(0, "handleDisconnect from %s", moduleName) // REMOVE
	if moduleName != "" {
		holder.EtpServer.Rooms().Leave(conn, service.Room.Module(moduleName))
		now := state.GenerateDate()
		module := entity.Module{
			Id:                 state.GenerateId(),
			Name:               moduleName,
			CreatedAt:          now,
			LastConnectedAt:    now,
			LastDisconnectedAt: now,
		}
		command := cluster.PrepareModuleDisconnectedCommand(module)
		_, _ = SyncApplyCommand(command)
	}

	service.Discovery.HandleDisconnect(conn.ID())
	service.Routes.HandleDisconnect(conn.ID())
	backend, ok := ExtractBackendDeclaration(conn)
	if ok {
		command := cluster.PrepareDeleteBackendDeclarationCommand(backend)
		_, _ = SyncApplyCommand(command)
	}
}

func (h *SocketEventHandler) handleError(_ etp.Conn, err error) {
	log.Warnf(codes.WebsocketError, "isp-etp: %v", err)
}

func NewSocketEventHandler(server etp.Server, store *store.Store) *SocketEventHandler {
	return &SocketEventHandler{
		server: server,
		store:  store,
	}
}
