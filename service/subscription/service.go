package subscription

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity"
	"isp-config-service/helpers"
	"isp-config-service/service/rqlite/db"
)

type ModuleRepo interface {
	GetByNames(ctx context.Context, names []string) ([]entity.Module, error)
	GetById(ctx context.Context, id string) (*entity.Module, error)
}

type BackendRepo interface {
	GetByModuleId(ctx context.Context, moduleId string) ([]entity.Backend, error)
	All(ctx context.Context) ([]entity.Backend, error)
}

type ConfigRepo interface {
	GetActive(ctx context.Context, moduleId string) (*entity.Config, error)
}

type Service struct {
	moduleRepo  ModuleRepo
	backendRepo BackendRepo
	configRepo  ConfigRepo
	rooms       *etp.Rooms
	logger      log.Logger
}

func NewService(
	moduleRepo ModuleRepo,
	backendRepo BackendRepo,
	configRepo ConfigRepo,
	rooms *etp.Rooms,
	logger log.Logger,
) Service {
	return Service{
		moduleRepo:  moduleRepo,
		backendRepo: backendRepo,
		configRepo:  configRepo,
		rooms:       rooms,
		logger:      logger,
	}
}

func (s Service) SubscribeToConfigChanges(conn *etp.Conn, moduleId string) {
	s.rooms.Join(conn, ConfigChangingRoom(moduleId))
}

func (s Service) NotifyConfigChanged(ctx context.Context, moduleId string) error {
	conns := s.rooms.ToBroadcast(ConfigChangingRoom(moduleId))
	if len(conns) == 0 {
		return nil
	}

	ctx = db.NoneConsistency().ToContext(ctx)
	config, err := s.configRepo.GetActive(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "get active config")
	}

	for _, conn := range conns {
		go func() {
			s.emitEvent(ctx, conn, cluster.ConfigSendConfigChanged, config.Data)
		}()
	}

	return nil
}

func (s Service) SubscribeToBackendsChanges(ctx context.Context, conn *etp.Conn, requiredModuleNames []string) error {
	ctx = db.NoneConsistency().ToContext(ctx)
	modules, err := s.moduleRepo.GetByNames(ctx, requiredModuleNames)
	if err != nil {
		return errors.WithMessage(err, "get modules by names")
	}

	roomsToJoin := make([]string, 0)
	for _, module := range modules {
		roomsToJoin = append(roomsToJoin, BackendsChangingRoom(module.Id))
		go func() {
			err := s.notifyBackendsChanged(ctx, module.Id, []*etp.Conn{conn})
			if err != nil {
				s.logger.Error(ctx, errors.WithMessage(err, "notify backends changed"))
			}
		}()
	}

	s.rooms.Join(conn, roomsToJoin...)

	return nil
}

func (s Service) NotifyBackendsChanged(ctx context.Context, moduleId string) error {
	conns := s.rooms.ToBroadcast(BackendsChangingRoom(moduleId))
	if len(conns) == 0 {
		return nil
	}
	return s.notifyBackendsChanged(ctx, moduleId, conns)
}

func (s Service) notifyBackendsChanged(
	ctx context.Context,
	moduleId string,
	conns []*etp.Conn,
) error {
	ctx = db.NoneConsistency().ToContext(ctx)
	module, err := s.moduleRepo.GetById(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "get module by id")
	}
	backends, err := s.backendRepo.GetByModuleId(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "get backends by module id")
	}

	addresses := make([]cluster.AddressConfiguration, 0)
	for _, backend := range backends {
		addr, err := helpers.SplitAddress(backend)
		if err != nil {
			return errors.WithMessage(err, "split address")
		}
		addresses = append(addresses, addr)
	}
	data, err := json.Marshal(addresses)
	if err != nil {
		return errors.WithMessage(err, "marshal addresses")
	}

	for _, conn := range conns {
		go func() {
			s.emitEvent(ctx, conn, cluster.ModuleConnectedEvent(module.Name), data)
		}()
	}

	return nil
}

func (s Service) SubscribeToRoutingChanges(ctx context.Context, conn *etp.Conn) error {
	err := s.notifyRoutingChanged(ctx, cluster.ConfigSendRoutesWhenConnected, []*etp.Conn{conn})
	if err != nil {
		return errors.WithMessage(err, "notify routing changed")
	}

	s.rooms.Join(conn, RoutingChangingRoom())

	return nil
}

func (s Service) NotifyRoutingChanged(ctx context.Context) error {
	conns := s.rooms.ToBroadcast(RoutingChangingRoom())
	if len(conns) == 0 {
		return nil
	}
	return s.notifyRoutingChanged(ctx, cluster.ConfigSendRoutesChanged, conns)
}

func (s Service) notifyRoutingChanged(ctx context.Context, event string, conns []*etp.Conn) error {
	backends, err := s.backendRepo.All(ctx)
	if err != nil {
		return errors.WithMessage(err, "get all backends")
	}

	routingConfig := cluster.RoutingConfig{}
	for _, backend := range backends {
		addr, err := helpers.SplitAddress(backend)
		if err != nil {
			return errors.WithMessage(err, "split address")
		}
		declaration := cluster.BackendDeclaration{
			ModuleName:      backend.ModuleName,
			Version:         backend.Version,
			LibVersion:      backend.LibVersion,
			Endpoints:       backend.Endpoints.Value,
			RequiredModules: backend.RequiredModules.Value,
			Address:         addr,
		}
		routingConfig = append(routingConfig, declaration)
	}
	data, err := json.Marshal(routingConfig)
	if err != nil {
		return errors.WithMessage(err, "marshal routing config")
	}

	for _, conn := range conns {
		go func() {
			s.emitEvent(ctx, conn, event, data)
		}()
	}

	return nil
}

const (
	emitEventTimeout = 5 * time.Second
)

func (s Service) emitEvent(
	ctx context.Context,
	conn *etp.Conn,
	event string,
	data []byte,
) {
	ctx, cancel := context.WithTimeout(ctx, emitEventTimeout)
	defer cancel()
	err := conn.Emit(ctx, event, data)
	if err != nil {
		err := errors.WithMessagef(
			err,
			"emit event '%s', to %s module, connId: %d",
			event, helpers.ModuleName(conn), conn.Id(),
		)
		s.logger.Error(ctx, err)
	}
}
