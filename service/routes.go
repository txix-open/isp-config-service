package service

import (
	"context"
	"encoding/json"

	"github.com/cenkalti/backoff/v4"
	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/store/state"
)

var Routes routesService

type routesService struct{}

func (rs *routesService) HandleDisconnect(connID string) {
	holder.EtpServer.Rooms().LeaveByConnId(connID, Room.RoutesSubscribers())
}

func (rs *routesService) SubscribeRoutes(conn etp.Conn, mesh state.ReadonlyMesh) {
	holder.EtpServer.Rooms().Join(conn, Room.RoutesSubscribers())
	routes := mesh.GetRoutes()
	bytes, err := json.Marshal(routes)
	if err != nil {
		panic(err)
	}

	go func() {
		err := rs.sendRoutes(conn, utils.ConfigSendRoutesWhenConnected, bytes)
		if err != nil {
			log.Errorf(codes.RoutesServiceSendRoutesError, "send routes %v", err)
		}
	}()
}

func (rs *routesService) BroadcastRoutes(mesh state.ReadonlyMesh) {
	routes := mesh.GetRoutes()
	bytes, err := json.Marshal(routes)
	if err != nil {
		panic(err)
	}

	go func() {
		err := holder.EtpServer.BroadcastToRoom(Room.RoutesSubscribers(), utils.ConfigSendRoutesChanged, bytes)
		if err != nil {
			log.Errorf(codes.RoutesServiceSendRoutesError, "broadcast routes %v", err)
		}
	}()
}

func (rs *routesService) sendRoutes(conn etp.Conn, event string, bytes []byte) error {
	bf := backoff.WithMaxRetries(backoff.NewConstantBackOff(messagesBackoffInterval), messagesBackoffMaxRetries)
	err := backoff.Retry(func() error {
		return conn.Emit(context.Background(), event, bytes)
	}, bf)
	if err != nil {
		return err
	}
	return nil
}
