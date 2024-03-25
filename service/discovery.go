package service

import (
	"encoding/json"
	"sync"

	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/store/state"
)

const (
	AllModules = "*"
)

var (
	Discovery = &discoveryService{
		subs: make(map[string][]string),
	}
)

type discoveryService struct {
	subs map[string][]string
	lock sync.RWMutex
}

func (ds *discoveryService) HandleDisconnect(connID string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	if rooms, ok := ds.subs[connID]; ok {
		holder.EtpServer.Rooms().LeaveByConnId(connID, rooms...)
		delete(ds.subs, connID)
	}
}

func (ds *discoveryService) Subscribe(conn etp.Conn, modules []string, mesh state.ReadonlyMesh) {
	if len(modules) == 0 {
		return
	}
	ds.lock.Lock()
	defer ds.lock.Unlock()
	eventsAddresses := make([][]structure.AddressConfiguration, 0, len(modules))
	rooms := make([]string, 0, len(modules))
	for _, module := range modules {
		room := Room.AddressListener(module)
		addressList := mesh.GetModuleAddresses(module)
		eventsAddresses = append(eventsAddresses, addressList)
		rooms = append(rooms, room)
	}
	ds.subs[conn.ID()] = rooms
	holder.EtpServer.Rooms().Join(conn, rooms...)

	go func(events []string, eventsAddresses [][]structure.AddressConfiguration, conn etp.Conn) {
		for i := range events {
			event := events[i]
			addressList := eventsAddresses[i]
			body, err := json.Marshal(addressList)
			if err != nil {
				panic(err)
			}
			err = EmitConnWithTimeout(conn, event, body)
			if err != nil {
				log.Errorf(codes.DiscoveryServiceSendModulesError, "send module connected to %s: %v", conn.RemoteAddr(), err)
			}
		}
	}(rooms, eventsAddresses, conn)
}

func (ds *discoveryService) BroadcastModuleAddresses(moduleName string, mesh state.ReadonlyMesh) {
	ds.lock.RLock()
	defer ds.lock.RUnlock()
	room := Room.AddressListener(moduleName)
	event := utils.ModuleConnected(moduleName)
	addressList := mesh.GetModuleAddresses(moduleName)
	go ds.broadcastModuleAddrList(room, event, addressList)
}

func (ds *discoveryService) BroadcastEvent(event cluster.BroadcastEvent) {
	eventName := event.Event
	payload := make([]byte, len(event.Payload))
	copy(payload, event.Payload) // TODO а тут точно нужно копирование ?

	if len(event.ModuleNames) == 1 && event.ModuleNames[0] == AllModules {
		go func() {
			conns := holder.EtpServer.Rooms().AllConns()
			for _, conn := range conns {
				err := EmitConnWithTimeout(conn, eventName, payload)
				if err != nil {
					log.WithMetadata(map[string]interface{}{
						"eventName":      eventName,
						"remote_address": conn.RemoteAddr(),
					}).Errorf(codes.DiscoveryServiceSendModulesError, "broadcast event to all modules: %v", err)
				}
			}
		}()
		return
	}

	rooms := make([]string, 0, len(event.ModuleNames))
	for _, moduleName := range event.ModuleNames {
		room := Room.Module(moduleName)
		rooms = append(rooms, room)
	}
	go func(rooms []string) {
		conns := holder.EtpServer.Rooms().ToBroadcast(rooms...)
		for _, conn := range conns {
			err := EmitConnWithTimeout(conn, eventName, payload)
			if err != nil {
				log.WithMetadata(map[string]interface{}{
					"eventName":      eventName,
					"remote_address": conn.RemoteAddr(),
				}).Errorf(codes.DiscoveryServiceSendModulesError, "broadcast event: %v", err)
			}
		}
	}(rooms)
}

func (ds *discoveryService) broadcastModuleAddrList(room string, event string, addressList []structure.AddressConfiguration) {
	body, err := json.Marshal(addressList)
	if err != nil {
		panic(err)
	}
	conns := holder.EtpServer.Rooms().ToBroadcast(room)
	for _, conn := range conns {
		err := EmitConnWithTimeout(conn, event, body)
		if err != nil {
			log.Errorf(codes.DiscoveryServiceSendModulesError, "broadcast module connected to %s: %v", conn.RemoteAddr(), err)
		}
	}
}
