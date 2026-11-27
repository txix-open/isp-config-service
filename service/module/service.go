package module

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/helpers"
	"maps"
	"sort"
	"strings"

	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/rc"
	"github.com/txix-open/isp-kit/rc/schema"

	"github.com/google/uuid"
	"github.com/txix-open/etp/v4"
	"github.com/txix-open/etp/v4/store"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/log"
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
	GetByModuleId(ctx context.Context, moduleId string) (*entity.ConfigSchema, error)
}

type ConfigRepo interface {
	Insert(ctx context.Context, cfg entity.Config) error
	GetActive(ctx context.Context, moduleId string) (*entity.Config, error)
}

type VariableRepo interface {
	GetVariablesByName(ctx context.Context, names map[string]struct{}) ([]entity.Variable, error)
}

type VariableService interface {
	RenderConfig(ctx context.Context, moduleId string, input []byte) ([]byte, error)
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
	variableRepo        VariableRepo
	configWriteDeny     bool
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
	variableRepo VariableRepo,
	configWriteDeny bool,
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
		variableRepo:        variableRepo,
		configWriteDeny:     configWriteDeny,
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

	if config == nil {
		config, err = s.storeInitialConfig(ctx, conn, moduleId, data)
		if err != nil {
			return err
		}
	}

	configData, err := s.variableService.RenderConfig(ctx, moduleId, config.Data)
	if err != nil {
		return errors.WithMessage(err, "render config")
	}
	s.emitter.Emit(ctx, conn, cluster.ConfigSendConfigWhenConnected, configData)

	s.subscriptionService.SubscribeToConfigChanges(conn, moduleId)

	return nil
}

func idByKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// FilterSecretParams checks the config against the secret write policy:
// ui-exposed fields must not reference SECRET variables, and sensitive fields
// must hold only a single reference to a SECRET variable. A no-op when the
// config write policy is disabled.
func (s Service) FilterSecretParams(ctx context.Context, config json.RawMessage, moduleId string) error {
	if !s.configWriteDeny {
		return nil
	}

	moduleSchema, err := s.configSchemaRepo.GetByModuleId(ctx, moduleId)
	if err != nil {
		return errors.WithMessage(err, "find module schema")
	}

	if moduleSchema == nil {
		return nil
	}

	uiPaths, sensitivePaths, err := policyPaths(moduleSchema.Data)
	if err != nil {
		return err
	}
	if len(uiPaths) == 0 && len(sensitivePaths) == 0 {
		return nil
	}

	placeholders, err := rc.PlaceHolderReader(config)
	if err != nil {
		return errors.WithMessage(err, "read config placeholders")
	}

	refsByPath, names, sensitiveSet := policyRefs(placeholders, uiPaths, sensitivePaths)
	variableTypes, err := s.referencedVariableTypes(ctx, names)
	if err != nil {
		return err
	}

	errorDetails := uiExposedViolations(refsByPath, variableTypes)
	if len(sensitivePaths) > 0 {
		sensitiveDetails, err := sensitiveViolations(config, sensitiveSet, sensitivePaths, variableTypes)
		if err != nil {
			return err
		}
		maps.Copy(errorDetails, sensitiveDetails)
	}

	if len(errorDetails) == 0 {
		return nil
	}

	return errors.WithMessage(domain.NewConfigValidationError(errorDetails), "config violates the secret policy")
}

// storeInitialConfig checks the secret write policy and persists the module's
// initial config with its variable links. A violating default config is never
// persisted on the first module connection.
func (s Service) storeInitialConfig(ctx context.Context, conn *etp.Conn, moduleId string, data cluster.ConfigData) (*entity.Config, error) {
	err := s.FilterSecretParams(ctx, data.Config, moduleId)
	if err != nil {
		return nil, errors.WithMessage(err, "filter secret params in store")
	}

	initialConfig := entity.Config{
		Id:       idByKey(moduleId),
		Name:     helpers.ModuleName(conn),
		ModuleId: moduleId,
		Data:     data.Config,
		Version:  1,
		Active:   xtypes.Bool(true),
	}

	result, err := s.variableService.ExtractVariables(ctx, initialConfig.Id, initialConfig.Data)
	if err != nil {
		return nil, errors.WithMessage(err, "extract variables from initial config")
	}
	if len(result.AbsentVariables) > 0 {
		return nil, errors.Errorf("missing variables in initial config: [%s]", strings.Join(result.AbsentVariables, ","))
	}

	err = s.configRepo.Insert(ctx, initialConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "insert config in store")
	}
	err = s.variableService.SaveVariableLinks(ctx, initialConfig.Id, result.VariableLinks)
	if err != nil {
		return nil, errors.WithMessage(err, "save variable links in store")
	}

	return &initialConfig, nil
}

// referencedVariableTypes resolves the types of the variables referenced on
// the policy paths.
func (s Service) referencedVariableTypes(ctx context.Context, names map[string]struct{}) (map[string]entity.VariableType, error) {
	variableTypes := make(map[string]entity.VariableType, len(names))
	if len(names) == 0 {
		return variableTypes, nil
	}

	variables, err := s.variableRepo.GetVariablesByName(ctx, names)
	if err != nil {
		return nil, errors.WithMessage(err, "get variables")
	}
	for _, variable := range variables {
		variableTypes[variable.Name] = variable.Type
	}

	return variableTypes, nil
}

// policyPaths reads the ui-exposed and sensitive leaf paths from the schema.
func policyPaths(schemaData []byte) ([]string, []string, error) {
	uiPaths, err := schema.UIExposedPaths(schemaData)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "read schema params")
	}
	sensitivePaths, err := schema.SensitivePaths(schemaData)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "read schema params")
	}
	return uiPaths, sensitivePaths, nil
}

// policyRefs collects the ui-exposed references (kept per path for the
// report), the set of all referenced variable names and the sensitive paths.
func policyRefs(placeholders map[string][]string, uiPaths, sensitivePaths []string) (map[string][]string, map[string]struct{}, map[string]struct{}) {
	refsByPath := make(map[string][]string)
	names := make(map[string]struct{})
	sensitiveSet := make(map[string]struct{}, len(sensitivePaths))
	for _, path := range uiPaths {
		for _, name := range placeholders[path] {
			names[name] = struct{}{}
			refsByPath[path] = append(refsByPath[path], name)
		}
	}
	for _, path := range sensitivePaths {
		sensitiveSet[path] = struct{}{}
		for _, name := range placeholders[path] {
			names[name] = struct{}{}
		}
	}
	return refsByPath, names, sensitiveSet
}

// uiExposedViolations reports the ui-exposed paths that reference SECRET
// variables.
func uiExposedViolations(refsByPath map[string][]string, variableTypes map[string]entity.VariableType) map[string]string {
	details := make(map[string]string)
	for path, refs := range refsByPath {
		secrets := make([]string, 0, len(refs))
		for _, name := range refs {
			if variableTypes[name] == entity.SecretType {
				secrets = append(secrets, name)
			}
		}
		if len(secrets) == 0 {
			continue
		}
		sort.Strings(secrets)
		plural := "переменную"
		if len(secrets) > 1 {
			plural = "переменные"
		}
		message := fmt.Sprintf("UI-экспортируемое поле не должно ссылаться на %s с типом SECRET: %s", plural, strings.Join(secrets, ", "))
		details[policyFieldPath(path)] = message
	}
	return details
}

// sensitiveViolations reports the sensitive paths whose value is not a single
// reference to a SECRET variable. Composite JSON strings are report-only and
// never rejected.
func sensitiveViolations(
	config json.RawMessage,
	sensitiveSet map[string]struct{},
	sensitivePaths []string,
	variableTypes map[string]entity.VariableType,
) (map[string]string, error) {
	details := make(map[string]string)

	values, err := leafValuesAt(config, sensitiveSet)
	if err != nil {
		return nil, err
	}

	for _, path := range sensitivePaths {
		value, exists := values[path]
		if !exists || value == "" {
			continue
		}
		if rc.IsCompositeJSONString(value) {
			continue
		}
		if name, ok := rc.SingleVariableRef(value); ok {
			if t, known := variableTypes[name]; known && t != entity.SecretType {
				message := fmt.Sprintf("sensitive-поле должно ссылаться только на переменные с типом SECRET: переменная %s имеет тип %s", name, t)
				details[policyFieldPath(path)] = message
			}
			continue
		}
		details[policyFieldPath(path)] = "plain value in sensitive field"
	}

	return details, nil
}

// leafValuesAt collects the raw values of the config's string leaves located
// at the given canonical paths.
func leafValuesAt(data []byte, paths map[string]struct{}) (map[string]string, error) {
	values := make(map[string]string, len(paths))
	_, err := rc.TransformStringLeaves(data, func(path, value string) string {
		if _, want := paths[path]; want {
			values[path] = value
		}
		return value
	})
	if err != nil {
		return nil, errors.WithMessage(err, "read config leaf values")
	}
	return values, nil
}

func policyFieldPath(canonical string) string {
	path := strings.TrimPrefix(canonical, "$.")
	return strings.ReplaceAll(path, "[]", ".0")
}
