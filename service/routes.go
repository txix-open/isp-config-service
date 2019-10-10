package service

import (
	"encoding/json"
	"github.com/cenkalti/backoff"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/store/state"
	"isp-config-service/ws"
	"time"
)

const (
	RoutesSubscribersRoom = "__routesSubscribers"
)

var RoutesService routesService

type routesService struct{}

func (rs *routesService) HandleDisconnect(connId string) {
	holder.Socket.Rooms().LeaveByConnId(connId, RoutesSubscribersRoom)
}

func (rs *routesService) SubscribeRoutes(conn ws.Conn, state state.ReadState) {
	holder.Socket.Rooms().Join(conn, RoutesSubscribersRoom)
	routes := state.GetRoutes()
	err := rs.sendRoutes(conn, utils.ConfigSendRoutesWhenConnected, routes)
	if err != nil {
		log.Errorf(codes.RoutesServiceSendRoutesError, "send routes %v", err)
	}

}

func (rs *routesService) BroadcastRoutes(state state.ReadState) {
	routes := state.GetRoutes()
	err := rs.broadcastRoutes(utils.ConfigSendRoutesChanged, routes)
	if err != nil {
		log.Errorf(codes.RoutesServiceSendRoutesError, "broadcast routes %v", err)
	}

}

func (rs *routesService) broadcastRoutes(event string, routes structure.RoutingConfig) error {
	if bytes, err := json.Marshal(routes); err != nil {
		return err
	} else {
		err = holder.Socket.Broadcast(RoutesSubscribersRoom, event, string(bytes))
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *routesService) sendRoutes(conn ws.Conn, event string, routes structure.RoutingConfig) error {
	if bytes, err := json.Marshal(routes); err != nil {
		return err
	} else {
		bf := backoff.WithMaxRetries(backoff.NewConstantBackOff(100*time.Millisecond), 3)
		err := backoff.Retry(func() error {
			return conn.Emit(event, string(bytes))
		}, bf)
		if err != nil {
			return err
		}
	}
	return nil
}
