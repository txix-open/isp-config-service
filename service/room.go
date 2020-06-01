package service

import "github.com/integration-system/isp-lib/v2/utils"

const (
	configWatchersRoomSuffix = "_config"
	routesSubscribersRoom    = "__routesSubscribers"
)

var (
	Room = roomService{}
)

type roomService struct{}

func (s roomService) Module(moduleName string) string {
	return moduleName + configWatchersRoomSuffix
}

func (s roomService) RoutesSubscribers() string {
	return routesSubscribersRoom
}

func (s roomService) AddressListener(moduleName string) string {
	return utils.ModuleConnected(moduleName)
}
