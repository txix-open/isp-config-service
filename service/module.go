package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/txix-open/etp/v3"
	"github.com/txix-open/etp/v3/store"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity"
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
		disconnected time.Time,
	) error
}

type BackendRepo interface {
	Upsert(ctx context.Context, backend entity.Backend) error
	Delete(ctx context.Context, moduleId string, address string) error
}

type Module struct {
	moduleRepo  ModuleRepo
	backendRepo BackendRepo
	logger      log.Logger
}

func NewModule(
	moduleRepo ModuleRepo,
	backendRepo BackendRepo,
	logger log.Logger,
) Module {
	return Module{
		moduleRepo:  moduleRepo,
		backendRepo: backendRepo,
		logger:      logger,
	}
}

func (s Module) OnConnect(conn *etp.Conn, moduleName string) error {
	now := time.Now().UTC()
	module := entity.Module{
		Id:              uuid.NewString(),
		Name:            moduleName,
		LastConnectedAt: &now,
		CreatedAt:       now,
	}
	moduleId, err := s.moduleRepo.Upsert(context.Background(), module)
	if err != nil {
		return errors.WithMessage(err, "upsert module in store")
	}

	conn.Data().Set(moduleIdKey, moduleId)

	return nil
}

func (s Module) OnDisconnect(
	conn *etp.Conn,
	moduleName string,
	isNormalClose bool,
	err error,
) error {
	if isNormalClose {
		s.logger.Info(context.Background(), fmt.Sprintf("module '%s' disconnected", moduleName))
	} else {
		message := errors.WithMessagef(
			err,
			"module '%s' unexpectedly disconnected",
			moduleName,
		)
		s.logger.Error(context.Background(), message)
	}

	moduleId, _ := store.Get[string](conn.Data(), moduleIdKey)
	if moduleId != "" {
		err = s.moduleRepo.SetDisconnectedAt(context.Background(), moduleName, time.Now().UTC())
		if err != nil {
			return errors.WithMessage(err, "update disconnected time in store")
		}
	}

	backend, _ := store.Get[entity.Backend](conn.Data(), backendKey)
	if backend.ModuleId != "" {
		err = s.backendRepo.Delete(context.Background(), backend.ModuleId, backend.Address)
		if err != nil {
			return errors.WithMessage(err, "delete backend in store")
		}
	}

	return nil
}

func (s Module) OnError(conn *etp.Conn, moduleName string, err error) {
	err = errors.WithMessagef(
		err,
		"unexpected error in communication, module: '%s'",
		moduleName,
	)
	s.logger.Error(context.Background(), err)
}

func (s Module) OnModuleReady(
	ctx context.Context,
	conn *etp.Conn,
	backend cluster.BackendDeclaration,
) error {
	moduleId, err := store.Get[string](conn.Data(), moduleIdKey)
	if err != nil {
		return errors.WithMessage(err, "resolve module id")
	}

	backend := entity.Backend{
		ModuleId:        moduleId,
		Address:         fmt.Sprintf("%s:%s", backend.Address.IP, backend.Address.Port),
		Version:         backend.Version,
		LibVersion:      backend.LibVersion,
		Endpoints:       nil,
		RequiredModules: nil,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
	}
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
	return nil
}
