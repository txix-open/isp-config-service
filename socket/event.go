package socket

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/integration-system/isp-lib/bootstrap"
	"github.com/integration-system/isp-lib/structure"

	"isp-config-service/service"

	"github.com/googollee/go-socket.io"
	"github.com/integration-system/isp-lib/config/schema"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
	"isp-config-service/entity"
	"isp-config-service/model"
)

const (
	routesSubscribers = ":routes_subs"
)

var (
	okMsg    = "ok"
	errorMsg = "error"
)

func RoutesSubscribersRoom(instanceUuid string) string {
	return instanceUuid + routesSubscribers
}

func onError(so socketio.Socket, err error) {
	logger.Warnf("error: %s", err)
}

func onDisconnect(so socketio.Socket) {
	logger.Debugf("onDisconnect: %s", so.Id())
	instanceUuid, moduleName, err := utils.ParseParameters(so.Request().URL.RawQuery)
	if err != nil {
		logger.Warn(err)
		return
	}

	must(so.Leave(instanceUuid + moduleName))
	must(so.Leave(instanceUuid))
	must(so.Leave(RoutesSubscribersRoom(instanceUuid)))
	must(service.Disonnect(instanceUuid, moduleName))

	discoverer := getOrRegisterDiscoverer(instanceUuid)
	discoverer.OnDisconnection(so.Id())
	discoverer.BroadcastModuleAddresses(moduleName)
	if backendRoutes.DeleteRoute(so.Id()) {
		BroadcastRoutesSnapshot(instanceUuid)
	}
}

func onConnect(so socketio.Socket) {
	logger.Debugf("onConnect: %s", so.Id())

	instanceUuid, moduleName, err := utils.ParseParameters(so.Request().URL.RawQuery)
	if err != nil {
		_ = so.Emit(utils.ConfigError, err.Error())
		return
	}

	config, err := service.NewConnection(instanceUuid, moduleName)
	if err != nil {
		_ = so.Emit(utils.ConfigError, err.Error())
		return
	}

	must(so.Join(instanceUuid + moduleName))
	must(so.Join(instanceUuid))
	_ = so.Emit(
		utils.ConfigSendConfigWhenConnected, config.Data.ToJSON(), func(so socketio.Socket, data string) {
			logger.Debug("Client ACK with data: ", data)
		},
	)
}

func onReceivedModuleRequirements(so socketio.Socket, msg string) {
	logger.Debugf("onReceivedModuleRequirements: %s %s", so.Id(), msg)

	instanceUuid, _, err := utils.ParseParameters(so.Request().URL.RawQuery)
	if err != nil {
		_ = so.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		return
	}

	declaration := bootstrap.ModuleRequirements{}
	err = json.Unmarshal([]byte(msg), &declaration)
	if err != nil {
		_ = so.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		logger.Debugf("onReceivedModuleRequirements: %s, error parse json data: %s", so.Id(), err.Error())
		return
	}

	if declaration.RequireRoutes {
		must(so.Join(RoutesSubscribersRoom(instanceUuid)))
		SendRoutesSnapshot(so)
	}
	discoverer := getOrRegisterDiscoverer(instanceUuid)
	discoverer.Subscribe(declaration.RequiredModules, so)
}

func onReceiveModuleReady(so socketio.Socket, msg string) {
	logger.Debugf("onReceiveModuleReady: %s", so.Id())

	handleModuleDeclaration(so, msg)
}

func onReceiveRoutesUpdate(so socketio.Socket, msg string) {
	logger.Debugf("onReceiveRoutesUpdate: %s", so.Id())

	handleModuleDeclaration(so, msg)
}

func onReceiveRemoteConfigSchema(so socketio.Socket, msg string) string {
	logger.Debugf("onReceiveRemoteConfigSchema: %s %s", so.Id(), msg)

	instanceUuid, moduleName, err := utils.ParseParameters(so.Request().URL.RawQuery)
	if err != nil {
		logger.Warn(err)
		return errorMsg
	}

	module, err := model.ModulesRep.GetModulesByInstanceUuidAndName(instanceUuid, moduleName)
	if err != nil {
		logger.Warn(err)
		return errorMsg
	}
	if module == nil {
		logger.Warnf("Module '%s' in instance '%s' not found", moduleName, instanceUuid)
		return errorMsg
	}

	s := new(schema.ConfigSchema)
	if err := json.Unmarshal([]byte(msg), s); err != nil {
		logger.Warn(err)
		return errorMsg
	}

	res, err := model.SchemaRep.GetSchemasByModulesId([]int32{module.Id})
	if err != nil {
		logger.Warn(err)
		return errorMsg
	}
	if len(res) > 0 {
		schema := &res[0]
		schema.Schema = s.Schema
		schema.Version = s.Version
		if _, err := model.SchemaRep.UpdateConfigSchema(schema); err != nil {
			logger.Warn(err)
			return errorMsg
		}
		/*if s.Version > schema.Version {
			schema.Version = s.Version
			schema.Schema = s.Schema
			if _, err := model.SchemaRep.UpdateConfigSchema(schema); err != nil {
				logger.Warn(err)
			}
		}*/
	} else {
		cs := &entity.ConfigSchema{
			Version:  s.Version,
			Schema:   s.Schema,
			ModuleId: module.Id,
		}
		if _, err := model.SchemaRep.InsertConfigSchema(cs); err != nil {
			logger.Warn(err)
			return errorMsg
		}
	}

	if len(s.DefaultConfig) == 0 {
		return okMsg
	}

	configs, err := model.ConfigRep.GetConfigByInstanceUUIDAndModuleName(instanceUuid, moduleName)
	if err != nil {
		logger.Warn(err)
		return errorMsg
	}
	if configs != nil {
		return okMsg
	}
	config := entity.Config{
		ModuleId: module.Id,
		Name:     module.Name,
		Data:     s.DefaultConfig,
		Active:   true,
		Version:  1,
	}
	_, err = model.ConfigRep.CreateConfig(&config)
	if err != nil {
		logger.Warn(err)
		return errorMsg
	}
	_ = so.Emit(
		utils.ConfigSendConfigWhenConnected, config.Data.ToJSON(), func(so socketio.Socket, data string) {
			logger.Debug("Client ACK with data: ", data)
		},
	)

	return okMsg
}

func handleModuleDeclaration(so socketio.Socket, msg string) {
	instanceUuid, moduleName, err := utils.ParseParameters(so.Request().URL.RawQuery)
	if err != nil {
		_ = so.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		return
	}

	declaration := structure.BackendDeclaration{}
	err = json.Unmarshal([]byte(msg), &declaration)
	if err != nil {
		_ = so.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		logger.Warnf("handleModuleDeclaration: %s, error parse json data: %s", so.Id(), err.Error())
		return
	}

	_, err = govalidator.ValidateStruct(declaration)
	if err != nil {
		errors := govalidator.ErrorsByField(err)
		_ = so.Emit(utils.ErrorConnection, map[string]map[string]string{"error": errors})
		logger.Warnf("SOCKET ROUTES ERROR, handleModuleDeclaration: %s, error validate routes data: %s", so.Id(), errors)
	} else if backendRoutes.AddAddressOrUpdate(so.Id(), declaration) {
		BroadcastRoutesSnapshot(instanceUuid)

		discoverer := getOrRegisterDiscoverer(instanceUuid)
		discoverer.BroadcastModuleAddresses(moduleName)
	}
}

func onRequestConfig(so socketio.Socket) {
	logger.Debugf("onRequestConfig: %s", so.Id())

	instanceUuid, moduleName, err := utils.ParseParameters(so.Request().URL.RawQuery)
	if err != nil {
		_ = so.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		return
	}
	config, err := service.GetConfig(instanceUuid, moduleName)
	if err != nil {
		_ = so.Emit(utils.ErrorConnection, map[string]string{"error": err.Error()})
		return
	}

	_ = so.Emit(
		utils.ConfigSendConfigOnRequest,
		config.Data.ToJSON(),
		func(so socketio.Socket, data string) {
			logger.Debug("Client ACK with data: ", data)
		},
	)
}
