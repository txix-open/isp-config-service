package module

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/txix-open/etp/v3"
	"github.com/txix-open/etp/v3/store"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/helpers"
)

const (
	moduleIdKey = "moduleId"
	backendKey  = "backend"
)

type Repo interface {
	Upsert(ctx context.Context, module entity.Module) (string, error)
	SetDisconnectedAtNow(
		ctx context.Context,
		moduleId string,
	) error
}

type BackendRepo interface {
	Upsert(ctx context.Context, backend entity.Backend) error
	Delete(ctx context.Context, moduleId string, address string) error
}

type EventRepo interface {
	Insert(ctx context.Context, event entity.Event) error
}

type ConfigSchemaRepo interface {
	Upsert(ctx context.Context, schema entity.ConfigSchema) error
}

type ConfigRepo interface {
	Insert(ctx context.Context, cfg entity.Config) error
	GetActive(ctx context.Context, moduleId string) (*entity.Config, error)
}

type SubscriptionService interface {
	SubscribeToConfigChanges(conn *etp.Conn, moduleId string)
	SubscribeToBackendsChanges(ctx context.Context, conn *etp.Conn, requiredModuleNames []string) error
	SubscribeToRoutingChanges(ctx context.Context, conn *etp.Conn) error
}

type Service struct {
	moduleRepo          Repo
	backendRepo         BackendRepo
	eventRepo           EventRepo
	configRepo          ConfigRepo
	configSchemaRepo    ConfigSchemaRepo
	subscriptionService SubscriptionService
	emitter             Emitter
	logger              log.Logger
}

func NewService(
	moduleRepo Repo,
	backendRepo BackendRepo,
	eventRepo EventRepo,
	configRepo ConfigRepo,
	configSchemaRepo ConfigSchemaRepo,
	subscriptionService SubscriptionService,
	emitter Emitter,
	logger log.Logger,
) Service {
	return Service{
		moduleRepo:          moduleRepo,
		backendRepo:         backendRepo,
		eventRepo:           eventRepo,
		configRepo:          configRepo,
		configSchemaRepo:    configSchemaRepo,
		subscriptionService: subscriptionService,
		emitter:             emitter,
		logger:              logger,
	}
}

func (s Service) OnConnect(ctx context.Context, conn *etp.Conn, moduleName string) error {
	s.logger.Info(ctx, "module connected", helpers.LogFields(conn)...)

	module := entity.Module{
		Id:   uuid.NewString(),
		Name: moduleName,
	}
	moduleId, err := s.moduleRepo.Upsert(ctx, module)
	if err != nil {
		return errors.WithMessage(err, "upsert module in store")
	}

	conn.Data().Set(moduleIdKey, moduleId)

	return nil
}

func (s Service) OnDisconnect(
	ctx context.Context,
	conn *etp.Conn,
	moduleName string,
	isNormalClose bool,
	err error,
) error {
	if isNormalClose {
		s.logger.Info(
			ctx,
			"module disconnected",
			helpers.LogFields(conn)...,
		)
	} else {
		message := errors.WithMessage(
			err,
			"module unexpectedly disconnected",
		)
		s.logger.Error(ctx, message, helpers.LogFields(conn)...)
	}

	moduleId, _ := store.Get[string](conn.Data(), moduleIdKey)
	if moduleId != "" {
		err = s.moduleRepo.SetDisconnectedAtNow(ctx, moduleName)
		if err != nil {
			return errors.WithMessage(err, "update disconnected time in store")
		}
	}

	backend, _ := store.Get[entity.Backend](conn.Data(), backendKey)
	if backend.ModuleId != "" {
		err = s.backendRepo.Delete(ctx, backend.ModuleId, backend.Address)
		if err != nil {
			return errors.WithMessage(err, "delete backend in store")
		}

		event := entity.NewEvent(entity.EventPayload{
			ModuleDisconnected: &entity.PayloadModuleDisconnected{
				ModuleId: moduleId,
			},
		})
		err = s.eventRepo.Insert(ctx, event)
		if err != nil {
			return errors.WithMessage(err, "insert event in store")
		}
	}

	return nil
}

func (s Service) OnError(ctx context.Context, conn *etp.Conn, err error) {
	err = errors.WithMessage(
		err,
		"unexpected error in communication",
	)
	s.logger.Error(ctx, err, helpers.LogFields(conn)...)
}

func (s Service) OnModuleReady(
	ctx context.Context,
	conn *etp.Conn,
	declaration cluster.BackendDeclaration,
) error {
	moduleId, err := store.Get[string](conn.Data(), moduleIdKey)
	if err != nil {
		return errors.WithMessage(err, "resolve module id")
	}

	backend := entity.Backend{
		ModuleId:        moduleId,
		Address:         fmt.Sprintf("%s:%s", declaration.Address.IP, declaration.Address.Port),
		Version:         declaration.Version,
		LibVersion:      declaration.LibVersion,
		ModuleName:      declaration.ModuleName,
		Endpoints:       xtypes.Json[[]cluster.EndpointDescriptor]{Value: declaration.Endpoints},
		RequiredModules: xtypes.Json[[]cluster.ModuleDependency]{Value: declaration.RequiredModules},
	}
	err = s.backendRepo.Upsert(ctx, backend)
	if err != nil {
		return errors.WithMessage(err, "upsert backend in store")
	}

	conn.Data().Set(backendKey, backend)

	event := entity.NewEvent(entity.EventPayload{
		ModuleReady: &entity.PayloadModuleReady{
			ModuleId: moduleId,
		},
	})
	err = s.eventRepo.Insert(ctx, event)
	if err != nil {
		return errors.WithMessage(err, "insert event in store")
	}

	return nil
}

func (s Service) OnModuleRequirements(
	ctx context.Context,
	conn *etp.Conn,
	requirements cluster.ModuleRequirements,
) error {
	if len(requirements.RequiredModules) > 0 {
		err := s.subscriptionService.SubscribeToBackendsChanges(ctx, conn, requirements.RequiredModules)
		if err != nil {
			return errors.WithMessage(err, "subscribe to required modules")
		}
	}

	if requirements.RequireRoutes {
		err := s.subscriptionService.SubscribeToRoutingChanges(ctx, conn)
		if err != nil {
			return errors.WithMessage(err, "subscribe to routing changes")
		}
	}

	return nil
}

func (s Service) OnModuleConfigSchema(
	ctx context.Context,
	conn *etp.Conn,
	data cluster.ConfigData,
) error {
	moduleId, err := store.Get[string](conn.Data(), moduleIdKey)
	if err != nil {
		return errors.WithMessage(err, "resolve module id")
	}

	config, err := s.configRepo.GetActive(ctx, moduleId)
	if config == nil {
		md5Sum := md5.Sum([]byte(moduleId))
		initialConfigId := hex.EncodeToString(md5Sum[:])
		initialConfig := entity.Config{
			Id:       initialConfigId,
			Name:     helpers.ModuleName(conn),
			ModuleId: moduleId,
			Data:     data.Config,
			Version:  1,
			Active:   xtypes.Bool(true),
		}
		err := s.configRepo.Insert(ctx, initialConfig)
		if err != nil {
			return errors.WithMessage(err, "insert config in store")
		}
		config = &initialConfig
	}
	if err != nil {
		return errors.WithMessage(err, "get active config")
	}

	s.emitter.Emit(ctx, conn, cluster.ConfigSendConfigWhenConnected, config.Data)

	s.subscriptionService.SubscribeToConfigChanges(conn, moduleId)

	configSchema := entity.ConfigSchema{
		Id:            uuid.NewString(),
		ModuleId:      moduleId,
		Data:          data.Schema,
		ModuleVersion: data.Version,
	}
	err = s.configSchemaRepo.Upsert(ctx, configSchema)
	if err != nil {
		return errors.WithMessage(err, "upsert config schema in store")
	}

	return nil
}
