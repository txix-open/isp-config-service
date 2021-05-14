package service

import (
	"encoding/json"

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
	body, err := json.Marshal(routes)
	if err != nil {
		panic(err)
	}

	go func() {
		err := EmitConnWithTimeout(conn, utils.ConfigSendRoutesWhenConnected, body)
		if err != nil {
			log.Errorf(codes.RoutesServiceSendRoutesError, "send routes to %s: %v", conn.RemoteAddr(), err)
		}
	}()
}

func (rs *routesService) BroadcastRoutes(mesh state.ReadonlyMesh) {
	routes := mesh.GetRoutes()
	body, err := json.Marshal(routes)
	if err != nil {
		panic(err)
	}

	go func() {
		conns := holder.EtpServer.Rooms().ToBroadcast(Room.RoutesSubscribers())
		for _, conn := range conns {
			err := EmitConnWithTimeout(conn, utils.ConfigSendRoutesChanged, body)
			if err != nil {
				log.Errorf(codes.RoutesServiceSendRoutesError, "broadcast routes to %s: %v", conn.RemoteAddr(), err)
			}
		}
	}()
}
