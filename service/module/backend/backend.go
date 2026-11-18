package backend

import (
	"context"
	"fmt"
	"net/http"

	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/service/module/leader"

	"github.com/txix-open/etp/v4"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/log"
)

type Repo interface {
	Upsert(ctx context.Context, backend entity.Backend) error
	DeleteByWsConnectionIds(ctx context.Context, wsConnIds []string) error
	GetByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) ([]entity.Backend, error)
	DeleteByConfigServiceNodeId(ctx context.Context, configServiceNodeId string) (int, error)
	GetForLeaderElection(ctx context.Context, moduleId string) ([]entity.Backend, error)
}

type EventRepo interface {
	Insert(ctx context.Context, event entity.Event) error
}

type Backend struct {
	backendRepo Repo
	eventRepo   EventRepo
	nodeId      string
	rooms       *etp.Rooms
	logger      log.Logger
}

func NewBackend(
	backendRepo Repo,
	eventRepo EventRepo,
	nodeId string,
	rooms *etp.Rooms,
	logger log.Logger,
) Backend {
	return Backend{
		backendRepo: backendRepo,
		eventRepo:   eventRepo,
		nodeId:      nodeId,
		rooms:       rooms,
		logger:      logger,
	}
}

func (s Backend) Connect(
	ctx context.Context,
	connId string,
	moduleId string,
	declaration cluster.BackendDeclaration,
) (*entity.Backend, error) {
	var metricsAd *xtypes.Json[cluster.MetricsAutodiscovery]
	if declaration.MetricsAutodiscovery != nil {
		metricsAd = &xtypes.Json[cluster.MetricsAutodiscovery]{Value: *declaration.MetricsAutodiscovery}
	}

	if declaration.Transport == "" {
		declaration.Transport = cluster.GrpcTransport
	}
	if declaration.Transport == cluster.GrpcTransport {
		for i := range declaration.Endpoints {
			if declaration.Endpoints[i].HttpMethod == "" {
				declaration.Endpoints[i].HttpMethod = http.MethodPost
			}
		}
	}

	backend := entity.Backend{
		WsConnectionId:       connId,
		ModuleId:             moduleId,
		Address:              fmt.Sprintf("%s:%s", declaration.Address.IP, declaration.Address.Port),
		Version:              declaration.Version,
		LibVersion:           declaration.LibVersion,
		ModuleName:           declaration.ModuleName,
		Transport:            declaration.Transport,
		ModuleLeaderMode:     s.normalizeModuleLeaderMode(ctx, declaration.ModuleName, declaration.ModuleLeaderMode),
		ConfigServiceNodeId:  s.nodeId,
		Endpoints:            xtypes.Json[[]cluster.EndpointDescriptor]{Value: declaration.Endpoints},
		RequiredModules:      xtypes.Json[[]cluster.ModuleDependency]{Value: declaration.RequiredModules},
		CreatedAt:            xtypes.Time{},
		MetricsAutodiscovery: metricsAd,
	}

	previousLeader, err := s.leaderAddress(ctx, moduleId)
	if err != nil {
		return nil, errors.WithMessagef(err, "get leader of module %s before connect", moduleId)
	}

	err = s.backendRepo.Upsert(ctx, backend)
	if err != nil {
		return nil, errors.WithMessage(err, "upsert backend in store")
	}

	leaderAddress, err := s.leaderAddress(ctx, moduleId)
	if err != nil {
		return nil, errors.WithMessagef(err, "get leader of module %s after connect", moduleId)
	}

	event := entity.NewEvent(entity.EventPayload{
		ModuleReady: &entity.PayloadModuleReady{
			ModuleId: moduleId,
		},
	})
	err = s.eventRepo.Insert(ctx, event)
	if err != nil {
		return nil, errors.WithMessage(err, "insert event in store")
	}

	err = s.insertLeaderRecomputeEvent(ctx, entity.PayloadModuleLeaderRecompute{
		ModuleId:       moduleId,
		ModuleName:     backend.ModuleName,
		Reason:         entity.ModuleLeaderReasonReady,
		PreviousLeader: previousLeader,
		Leader:         leaderAddress,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "insert module leader recompute event")
	}

	return &backend, nil
}

func (s Backend) Disconnect(ctx context.Context, backend entity.Backend) error {
	previousLeader, err := s.leaderAddress(ctx, backend.ModuleId)
	if err != nil {
		return errors.WithMessagef(err, "get leader of module %s before disconnect", backend.ModuleId)
	}

	err = s.backendRepo.DeleteByWsConnectionIds(ctx, []string{backend.WsConnectionId})
	if err != nil {
		return errors.WithMessage(err, "delete backend in store")
	}

	leaderAddress, err := s.leaderAddress(ctx, backend.ModuleId)
	if err != nil {
		return errors.WithMessagef(err, "get leader of module %s after disconnect", backend.ModuleId)
	}

	err = s.insertDisconnectEvent(ctx, backend, previousLeader, leaderAddress)
	if err != nil {
		return errors.WithMessage(err, "insert disconnect event")
	}

	return nil
}

func (s Backend) ClearPhantomBackends(ctx context.Context) (int, error) {
	backends, err := s.backendRepo.GetByConfigServiceNodeId(ctx, s.nodeId)
	if err != nil {
		return 0, errors.WithMessagef(err, "get by config service node id = '%s'", s.nodeId)
	}

	toDelete := make([]string, 0)
	phantoms := make(map[string]entity.Backend, len(backends))
	for _, backend := range backends {
		_, ok := s.rooms.Get(backend.WsConnectionId)
		if ok {
			continue
		}

		toDelete = append(toDelete, backend.WsConnectionId)
		phantoms[backend.ModuleId] = backend
	}

	if len(toDelete) == 0 {
		return 0, nil
	}

	previousLeaders := make(map[string]string, len(phantoms))
	for moduleId := range phantoms {
		previousLeader, err := s.leaderAddress(ctx, moduleId)
		if err != nil {
			return 0, errors.WithMessagef(err, "get leader of module %s before phantom cleanup", moduleId)
		}
		previousLeaders[moduleId] = previousLeader
	}

	err = s.backendRepo.DeleteByWsConnectionIds(ctx, toDelete)
	if err != nil {
		return 0, errors.WithMessage(err, "delete backends by ws connection ids")
	}

	for moduleId, phantom := range phantoms {
		leaderAddress, err := s.leaderAddress(ctx, moduleId)
		if err != nil {
			return 0, errors.WithMessagef(err, "get leader of module %s after phantom cleanup", moduleId)
		}

		err = s.insertDisconnectEvent(ctx, phantom, previousLeaders[moduleId], leaderAddress)
		if err != nil {
			return 0, errors.WithMessage(err, "insert disconnect event")
		}
	}

	return len(toDelete), nil
}

func (s Backend) DeleteOwnBackends(ctx context.Context) {
	deleted, err := s.backendRepo.DeleteByConfigServiceNodeId(ctx, s.nodeId)
	if err != nil {
		s.logger.Error(ctx, errors.WithMessagef(err, "delete by config service node id = '%s'", s.nodeId))
		return
	}
	s.logger.Info(ctx, fmt.Sprintf("delete %d own backends", deleted))
}

func (s Backend) insertDisconnectEvent(
	ctx context.Context,
	backend entity.Backend,
	previousLeader string,
	leaderAddress string,
) error {
	event := entity.NewEvent(entity.EventPayload{
		ModuleDisconnected: &entity.PayloadModuleDisconnected{
			ModuleId: backend.ModuleId,
		},
	})
	err := s.eventRepo.Insert(ctx, event)
	if err != nil {
		return errors.WithMessage(err, "insert event in store")
	}

	err = s.insertLeaderRecomputeEvent(ctx, entity.PayloadModuleLeaderRecompute{
		ModuleId:       backend.ModuleId,
		ModuleName:     backend.ModuleName,
		Reason:         entity.ModuleLeaderReasonDisconnect,
		PreviousLeader: previousLeader,
		Leader:         leaderAddress,
	})
	if err != nil {
		return errors.WithMessage(err, "insert module leader recompute event")
	}

	return nil
}

func (s Backend) insertLeaderRecomputeEvent(
	ctx context.Context,
	recompute entity.PayloadModuleLeaderRecompute,
) error {
	event := entity.NewEvent(entity.EventPayload{ModuleLeaderRecompute: &recompute})
	err := s.eventRepo.Insert(ctx, event)
	if err != nil {
		return errors.WithMessage(err, "insert event in store")
	}
	return nil
}

func (s Backend) leaderAddress(ctx context.Context, moduleId string) (string, error) {
	backends, err := s.backendRepo.GetForLeaderElection(ctx, moduleId)
	if err != nil {
		return "", errors.WithMessagef(err, "get backends by module id = %s", moduleId)
	}

	election := leader.Derive(backends)
	if !election.Found {
		return "", nil
	}

	return election.Leader.Address, nil
}

func (s Backend) normalizeModuleLeaderMode(
	ctx context.Context,
	moduleName string,
	moduleLeaderMode string,
) string {
	switch moduleLeaderMode {
	case cluster.ModuleLeaderModeNone, cluster.ModuleLeaderModeRequestElection:
		return moduleLeaderMode
	}

	s.logger.Warn(
		ctx,
		"unsupported module leader mode, stored as none",
		log.String("moduleName", moduleName),
		log.String("moduleLeaderMode", moduleLeaderMode),
	)

	return cluster.ModuleLeaderModeNone
}
