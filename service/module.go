package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

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

type ModuleRepo interface {
	Upsert(ctx context.Context, module entity.Module) (string, error)
	SetDisconnectedAt(
		ctx context.Context,
		moduleId string,
		disconnected xtypes.Time,
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

type Module struct {
	moduleRepo       ModuleRepo
	backendRepo      BackendRepo
	eventRepo        EventRepo
	configRepo       ConfigRepo
	configSchemaRepo ConfigSchemaRepo
	logger           log.Logger
}

func NewModule(
	moduleRepo ModuleRepo,
	backendRepo BackendRepo,
	eventRepo EventRepo,
	configRepo ConfigRepo,
	configSchemaRepo ConfigSchemaRepo,
	logger log.Logger,
) Module {
	return Module{
		moduleRepo:       moduleRepo,
		backendRepo:      backendRepo,
		eventRepo:        eventRepo,
		configRepo:       configRepo,
		configSchemaRepo: configSchemaRepo,
		logger:           logger,
	}
}

func (s Module) OnConnect(ctx context.Context, conn *etp.Conn, moduleName string) error {
	now := now()
	module := entity.Module{
		Id:              uuid.NewString(),
		Name:            moduleName,
		LastConnectedAt: &xtypes.Time{Value: now},
	}
	moduleId, err := s.moduleRepo.Upsert(ctx, module)
	if err != nil {
		return errors.WithMessage(err, "upsert module in store")
	}

	conn.Data().Set(moduleIdKey, moduleId)

	return nil
}

func (s Module) OnDisconnect(
	ctx context.Context,
	conn *etp.Conn,
	moduleName string,
	isNormalClose bool,
	err error,
) error {
	if isNormalClose {
		s.logger.Info(ctx, fmt.Sprintf("module '%s' disconnected", moduleName))
	} else {
		message := errors.WithMessagef(
			err,
			"module '%s' unexpectedly disconnected",
			moduleName,
		)
		s.logger.Error(ctx, message)
	}

	now := now()
	moduleId, _ := store.Get[string](conn.Data(), moduleIdKey)
	if moduleId != "" {
		err = s.moduleRepo.SetDisconnectedAt(ctx, moduleName, xtypes.Time{Value: now})
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

		event := entity.NewEvent(entity.ModuleDisconnected, entity.EventPayload{
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

func (s Module) OnError(ctx context.Context, conn *etp.Conn, moduleName string, err error) {
	err = errors.WithMessagef(
		err,
		"unexpected error in communication, module: '%s'",
		moduleName,
	)
	s.logger.Error(ctx, err)
}

func (s Module) OnModuleReady(
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
		Endpoints:       xtypes.Json[[]cluster.EndpointDescriptor]{Value: declaration.Endpoints},
		RequiredModules: xtypes.Json[[]cluster.ModuleDependency]{Value: declaration.RequiredModules},
	}
	err = s.backendRepo.Upsert(ctx, backend)
	if err != nil {
		return errors.WithMessage(err, "upsert backend in store")
	}

	conn.Data().Set(backendKey, backend)

	event := entity.NewEvent(entity.ModuleReady, entity.EventPayload{
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

func (s Module) OnModuleRequirements(
	ctx context.Context,
	conn *etp.Conn,
	requirements cluster.ModuleRequirements,
) error {
	return nil
}

func (s Module) OnModuleConfigSchema(
	ctx context.Context,
	conn *etp.Conn,
	data cluster.ConfigData,
) error {
	moduleId, err := store.Get[string](conn.Data(), moduleIdKey)
	if err != nil {
		return errors.WithMessage(err, "resolve module id")
	}

	config, err := s.configRepo.GetActive(ctx, moduleId)
	if errors.Is(err, entity.ErrNoActiveConfig) {
		md5Sum := md5.Sum([]byte(moduleId))
		initialConfigId := hex.EncodeToString(md5Sum[:])
		initialConfig := entity.Config{
			Id:       initialConfigId,
			Name:     helpers.ModuleName(conn),
			ModuleId: moduleId,
			Data:     data.Config,
			Version:  1,
			Active:   xtypes.Bool{Value: true},
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

	err = conn.Emit(ctx, cluster.ConfigSendConfigWhenConnected, config.Data)
	if err != nil {
		return errors.WithMessage(err, "send event with config")
	}

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

func now() time.Time {
	return time.Now().UTC()
}
