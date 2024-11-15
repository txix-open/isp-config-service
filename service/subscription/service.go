package subscription

import (
	"context"
	"github.com/txix-open/etp/v4"

	"github.com/pkg/errors"
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

type Emitter interface {
	Emit(ctx context.Context, conn *etp.Conn, event string, data []byte)
}

type Service struct {
	moduleRepo  ModuleRepo
	backendRepo BackendRepo
	configRepo  ConfigRepo
	rooms       *etp.Rooms
	emitter     Emitter
	logger      log.Logger
}

func NewService(
	moduleRepo ModuleRepo,
	backendRepo BackendRepo,
	configRepo ConfigRepo,
	rooms *etp.Rooms,
	emitter Emitter,
	logger log.Logger,
) Service {
	return Service{
		moduleRepo:  moduleRepo,
		backendRepo: backendRepo,
		configRepo:  configRepo,
		rooms:       rooms,
		emitter:     emitter,
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
			s.emitter.Emit(ctx, conn, cluster.ConfigSendConfigChanged, config.Data)
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
	moduleByName := make(map[string]entity.Module)
	for _, module := range modules {
		moduleByName[module.Name] = module
	}

	roomsToJoin := make([]string, 0)
	conns := []*etp.Conn{conn}
	for _, moduleName := range requiredModuleNames {
		roomsToJoin = append(roomsToJoin, BackendsChangingRoom(moduleName))
		go func() {
			module, ok := moduleByName[moduleName]
			if !ok {
				s.emitModuleConnectedEvent(ctx, conns, moduleName, []cluster.AddressConfiguration{})
				return
			}

			err := s.notifyBackendsChanged(ctx, module.Id, moduleName, conns)
			if err != nil {
				s.logger.Error(ctx, errors.WithMessage(err, "notify backends changed"))
			}
		}()
	}

	s.rooms.Join(conn, roomsToJoin...)

	return nil
}

func (s Service) NotifyBackendsChanged(ctx context.Context, moduleId string) error {
	ctx = db.NoneConsistency().ToContext(ctx)
	module, err := s.moduleRepo.GetById(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "get module by id")
	}
	if module == nil {
		return errors.Errorf("unknown module: %s", moduleId)
	}

	conns := s.rooms.ToBroadcast(BackendsChangingRoom(module.Name))
	if len(conns) == 0 {
		return nil
	}

	return s.notifyBackendsChanged(ctx, moduleId, module.Name, conns)
}

func (s Service) notifyBackendsChanged(
	ctx context.Context,
	moduleId string,
	moduleName string,
	conns []*etp.Conn,
) error {
	ctx = db.NoneConsistency().ToContext(ctx)
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

	s.emitModuleConnectedEvent(ctx, conns, moduleName, addresses)

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
	ctx = db.NoneConsistency().ToContext(ctx)
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
			s.emitter.Emit(ctx, conn, event, data)
		}()
	}

	return nil
}

func (s Service) emitModuleConnectedEvent(
	ctx context.Context,
	conns []*etp.Conn,
	moduleName string,
	addresses []cluster.AddressConfiguration,
) {
	data, _ := json.Marshal(addresses)

	for _, conn := range conns {
		go func() {
			s.emitter.Emit(ctx, conn, cluster.ModuleConnectedEvent(moduleName), data)
		}()
	}
}
