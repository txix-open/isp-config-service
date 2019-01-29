package socket

import (
	"encoding/json"
	"github.com/googollee/go-socket.io"
	"github.com/integration-system/isp-lib/config"
	"github.com/integration-system/isp-lib/logger"
	sck "github.com/integration-system/isp-lib/socket"
	str "github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"isp-config-service/conf"
	"isp-config-service/entity"
	"isp-config-service/model"
	"isp-config-service/service"
)

var (
	server                          *socketio.Server
	br                              *broadcaster
	socketRoomStats                 = sck.NewRoomStats()
	backendRoutes                   = &service.RouterTables{}
	ConfigServiceBackendDeclaration = &str.BackendDeclaration{}
)

func init() {
	newServer, err := socketio.NewServer(nil)
	if err != nil {
		logger.Fatal(err)
	}
	sessionStore := sck.NewSessionStore(sck.WithOnRemoveCallback(func(id string) {
		socketRoomStats.RemoveSocketConn(id)
	}))
	newServer.SetSessionManager(sessionStore)
	newServer.SetAdaptor(socketRoomStats)

	must(newServer.On("connection", onConnect))
	must(newServer.On("disconnection", onDisconnect))
	must(newServer.On("error", onError))
	must(newServer.On(utils.ModuleSendRequirements, onReceivedModuleRequirements))
	must(newServer.On(utils.ModuleReady, onReceiveModuleReady))
	must(newServer.On(utils.ModuleSendConfigSchema, onReceiveRemoteConfigSchema))
	must(newServer.On(utils.ModuleUpdateRoutes, onReceiveRoutesUpdate))

	server = newServer
	br = NewBroadcaster(server, 0)
}

func Get() *socketio.Server {
	return server
}

func GetRoutes() *service.RouterTables {
	return backendRoutes
}

func SetBackendMethods(backendConfig *str.BackendDeclaration) {
	ConfigServiceBackendDeclaration = backendConfig

	moduleName := config.Get().(*conf.Configuration).ModuleName
	backendRoutes.AddAddressOrUpdate(moduleName, *ConfigServiceBackendDeclaration)
	instances, err := model.InstanceRep.GetInstances(nil)
	if err != nil {
		logger.Warn("Error receiving instance from db", err)
	} else {
		for _, v := range instances {
			BroadcastRoutesSnapshot(v.Uuid)
		}
	}
}

func GetRoomsCount() map[string]map[string]int {
	return socketRoomStats.RoomsCount()
}

func GetModuleConnectionIdMapByInstanceId(instanceId string) map[string][]string {
	return socketRoomStats.GetModuleConnectionMap()[instanceId]
}

func GetConnectionById(instanceId, sockId string) (*sck.SocketConn, bool) {
	return socketRoomStats.GetConnection(instanceId, sockId)
}

func SendNewConfig(instanceUuid string, moduleName string, config *entity.Config) {
	moduleKey := instanceUuid + moduleName
	br.BroadcastConfig(moduleKey, config.Data.ToJSON(), nil)
}

func SendRoutesSnapshot(so socketio.Socket) {
	bytes, err := json.Marshal(backendRoutes.GetRoutes())
	if err != nil {
		logger.Warn("Error when serializing Backend Routes", err)
		return
	}

	if err := so.Emit(utils.ConfigSendRoutesWhenConnected, string(bytes)); err != nil {
		logger.Warn(err)
	}
}

func BroadcastRoutesSnapshot(instanceUuid string) {
	bytes, err := json.Marshal(backendRoutes.GetRoutes())
	if err != nil {
		logger.Warn("Error when serializing Backend Routes", err)
		return
	}

	br.BroadcastRoutes(RoutesSubscribersRoom(instanceUuid), string(bytes), func(data interface{}) {
		logger.Infof("Send %d routes", len(*backendRoutes.GetRoutes()))
		logger.Debugf("Send Routes: %v", *backendRoutes.GetRoutes())
	})
}
