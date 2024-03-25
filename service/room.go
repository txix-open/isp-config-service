package service

import (
	"context"
	"time"

	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/utils"
)

const (
	configWatchersRoomSuffix = "_config"
	routesSubscribersRoom    = "__routesSubscribers"
)

const wsWriteTimeout = time.Second

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

func EmitConnWithTimeout(conn etp.Conn, event string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), wsWriteTimeout)
	defer cancel()
	return conn.Emit(ctx, event, body)
}
