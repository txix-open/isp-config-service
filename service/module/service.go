package module

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/txix-open/etp/v4"
	"github.com/txix-open/etp/v4/store"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/helpers"
	"strings"
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

type BackendService interface {
	Connect(ctx context.Context, connId string, moduleId string, declaration cluster.BackendDeclaration) (*entity.Backend, error)
	Disconnect(ctx context.Context, backend entity.Backend) error
}

type ConfigSchemaRepo interface {
	Upsert(ctx context.Context, schema entity.ConfigSchema) error
}

type ConfigRepo interface {
	Insert(ctx context.Context, cfg entity.Config) error
	GetActive(ctx context.Context, moduleId string) (*entity.Config, error)
}

type VariableService interface {
	RenderConfig(ctx context.Context, input []byte) ([]byte, error)
	ExtractVariables(ctx context.Context, configId string, input []byte) (*domain.VariableExtractionResult, error)
	SaveVariableLinks(ctx context.Context, configId string, links []entity.ConfigHasVariable) error
}

type SubscriptionService interface {
	SubscribeToConfigChanges(conn *etp.Conn, moduleId string)
	SubscribeToBackendsChanges(ctx context.Context, conn *etp.Conn, requiredModuleNames []string) error
	SubscribeToRoutingChanges(ctx context.Context, conn *etp.Conn) error
}

type Service struct {
	moduleRepo          Repo
	backendService      BackendService
	configRepo          ConfigRepo
	configSchemaRepo    ConfigSchemaRepo
	subscriptionService SubscriptionService
	emitter             Emitter
	variableService     VariableService
	logger              log.Logger
}

func NewService(
	moduleRepo Repo,
	backendService BackendService,
	configRepo ConfigRepo,
	configSchemaRepo ConfigSchemaRepo,
	subscriptionService SubscriptionService,
	emitter Emitter,
	variableService VariableService,
	logger log.Logger,
) Service {
	return Service{
		moduleRepo:          moduleRepo,
		backendService:      backendService,
		configRepo:          configRepo,
		configSchemaRepo:    configSchemaRepo,
		subscriptionService: subscriptionService,
		emitter:             emitter,
		variableService:     variableService,
		logger:              logger,
	}
}

func (s Service) OnConnect(ctx context.Context, conn *etp.Conn, moduleName string) error {
	s.logger.Info(ctx, "module connected", helpers.LogFields(conn)...)

	moduleId := idByKey(moduleName)
	module := entity.Module{
		Id:   moduleId,
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
		err := s.backendService.Disconnect(ctx, backend)
		if err != nil {
			return errors.WithMessage(err, "disconnect")
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

	backend, err := s.backendService.Connect(ctx, conn.Id(), moduleId, declaration)
	if err != nil {
		return errors.WithMessage(err, "connect")
	}

	conn.Data().Set(backendKey, *backend)

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
	if err != nil {
		return errors.WithMessage(err, "get active config")
	}
	if config == nil {
		initialConfig := entity.Config{
			Id:       idByKey(moduleId),
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

		result, err := s.variableService.ExtractVariables(ctx, initialConfig.Id, initialConfig.Data)
		if err != nil {
			return errors.WithMessage(err, "extract and save variable links in store")
		}
		if len(result.AbsentVariables) > 0 {
			return errors.Errorf("missing variables in initial config: [%s]", strings.Join(result.AbsentVariables, ","))
		}
		err = s.variableService.SaveVariableLinks(ctx, initialConfig.Id, result.VariableLinks)
		if err != nil {
			return errors.WithMessage(err, "save variable links in store")
		}

		config = &initialConfig
	}

	configData, err := s.variableService.RenderConfig(ctx, config.Data)
	if err != nil {
		return errors.WithMessage(err, "render config")
	}
	s.emitter.Emit(ctx, conn, cluster.ConfigSendConfigWhenConnected, configData)

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

func idByKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
