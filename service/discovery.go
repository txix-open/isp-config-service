package service

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
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

	messagesBackoffInterval   = 100 * time.Millisecond
	messagesBackoffMaxRetries = 3
)

var (
	Discovery = newDiscoveryService()
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
			err := ds.sendAddrList(conn, event, addressList)
			if err != nil {
				log.Errorf(codes.DiscoveryServiceSendModulesError, "send module connected %v", err)
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
	go func(room, event string, addressList []structure.AddressConfiguration) {
		err := ds.broadcastAddrList(room, event, addressList)
		if err != nil {
			log.Errorf(codes.DiscoveryServiceSendModulesError, "broadcast module connected %v", err)
		}
	}(room, event, addressList)
}

func (ds *discoveryService) BroadcastEvent(event cluster.BroadcastEvent) {
	eventName := event.Event
	payload := make([]byte, len(event.Payload))
	copy(payload, event.Payload) //TODO а тут точно нужно копирование ?

	if len(event.ModuleNames) == 1 && event.ModuleNames[0] == AllModules {
		go func() {
			if err := holder.EtpServer.BroadcastToAll(eventName, payload); err != nil {
				err = errors.WithMessagef(err, "broadcast '%s' to all modules", eventName)
				log.Error(codes.DiscoveryServiceSendModulesError, err)
			}
		}()
	} else {
		for _, moduleName := range event.ModuleNames {
			go func(moduleName string) {
				room := Room.Module(moduleName)
				if err := holder.EtpServer.BroadcastToRoom(room, eventName, payload); err != nil {
					err = errors.WithMessagef(err, "broadcast '%s' to '%s'", eventName, moduleName)
					log.Error(codes.DiscoveryServiceSendModulesError, err)
				}
			}(moduleName)
		}
	}
}

func (ds *discoveryService) broadcastAddrList(room string, event string, addressList []structure.AddressConfiguration) error {
	bytes, err := json.Marshal(addressList)
	if err != nil {
		return err
	}
	err = holder.EtpServer.BroadcastToRoom(room, event, bytes)
	if err != nil {
		return err
	}
	return nil
}

func (ds *discoveryService) sendAddrList(conn etp.Conn, event string, addressList []structure.AddressConfiguration) error {
	bytes, err := json.Marshal(addressList)
	if err != nil {
		return err
	}
	bf := backoff.WithMaxRetries(backoff.NewConstantBackOff(messagesBackoffInterval), messagesBackoffMaxRetries)
	err = backoff.Retry(func() error {
		return conn.Emit(context.Background(), event, bytes)
	}, bf)
	if err != nil {
		return err
	}
	return nil
}

func newDiscoveryService() *discoveryService {
	return &discoveryService{
		subs: make(map[string][]string),
	}
}
