package service

import (
	"encoding/json"
	"github.com/cenkalti/backoff"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"isp-config-service/holder"
	"isp-config-service/store/state"
	"isp-config-service/ws"
	"time"
)

var RoutesService routesService

type routesService struct{}

func (rs *routesService) HandleDisconnect(connId string) {
	holder.Socket.Rooms().LeaveByConnId(connId, RoutesSubscribersRoom)
}

func (rs *routesService) SubscribeRoutes(conn ws.Conn, state state.ReadState) {
	holder.Socket.Rooms().Join(conn, RoutesSubscribersRoom)
	routes := state.GetRoutes()
	rs.sendRoutes(conn, utils.ConfigSendRoutesWhenConnected, routes)
}

func (rs *routesService) BroadcastRoutes(state state.ReadState) {
	routes := state.GetRoutes()
	rs.broadcastRoutes(utils.ConfigSendRoutesChanged, routes)
}

func (rs *routesService) broadcastRoutes(event string, routes structure.RoutingConfig) {
	if bytes, err := json.Marshal(routes); err != nil {
		logger.Warn(err)
	} else {
		err = holder.Socket.Broadcast(RoutesSubscribersRoom, event, string(bytes))
		if err != nil {
			logger.Error(err)
		}
	}
}

func (rs *routesService) sendRoutes(conn ws.Conn, event string, routes structure.RoutingConfig) {
	if bytes, err := json.Marshal(routes); err != nil {
		logger.Warn(err)
	} else {
		bf := backoff.WithMaxRetries(backoff.NewConstantBackOff(100*time.Millisecond), 3)
		err := backoff.Retry(func() error {
			return conn.Emit(event, string(bytes))
		}, bf)
		if err != nil {
			logger.Error(err)
		}
	}
}
