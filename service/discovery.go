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
	"strings"
	"sync"
	"time"
)

const (
	ModuleConnectionEventSuffix = "_" + utils.ModuleConnectionSuffix
)

var (
	DiscoveryService = NewDiscoveryService()
)

type discoveryService struct {
	subs map[string][]string
	lock sync.RWMutex
}

func (ds *discoveryService) HandleDisconnect(connId string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	if events, ok := ds.subs[connId]; ok {
		holder.Socket.Rooms().LeaveByConnId(connId, events...)
		delete(ds.subs, connId)
	}
}

func (ds *discoveryService) Subscribe(conn ws.Conn, events []string, mesh state.ReadonlyMesh) {
	if len(events) == 0 {
		return
	}
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ds.subs[conn.Id()] = events
	holder.Socket.Rooms().Join(conn, events...)
	eventsAddresses := make([][]structure.AddressConfiguration, 0, len(events))
	for _, event := range events {
		eventName := strings.TrimSuffix(event, ModuleConnectionEventSuffix)
		addressList := mesh.GetModuleAddresses(eventName)
		eventsAddresses = append(eventsAddresses, addressList)
	}
	go func(events []string, eventsAddresses [][]structure.AddressConfiguration, conn ws.Conn) {
		for i := range events {
			event := events[i]
			addressList := eventsAddresses[i]
			err := ds.sendAddrList(conn, event, addressList)
			if err != nil {
				log.Errorf(codes.DiscoveryServiceSendModulesError, "send module connected %v", err)
			}
		}
	}(events, eventsAddresses, conn)
}

func (ds *discoveryService) BroadcastModuleAddresses(moduleName string, mesh state.ReadonlyMesh) {
	ds.lock.RLock()
	defer ds.lock.RUnlock()
	event := utils.ModuleConnected(moduleName)
	addressList := mesh.GetModuleAddresses(moduleName)
	go func(moduleName, event string, addressList []structure.AddressConfiguration) {
		err := ds.broadcastAddrList(moduleName, event, addressList)
		if err != nil {
			log.Errorf(codes.DiscoveryServiceSendModulesError, "broadcast module connected %v", err)
		}
	}(moduleName, event, addressList)
}

func (ds *discoveryService) broadcastAddrList(moduleName string, event string, addressList []structure.AddressConfiguration) error {
	if bytes, err := json.Marshal(addressList); err != nil {
		return err
	} else {
		err = holder.Socket.Broadcast(moduleName, event, string(bytes))
		if err != nil {
			return err
		}
	}
	return nil
}

func (ds *discoveryService) sendAddrList(conn ws.Conn, event string, addressList []structure.AddressConfiguration) error {
	if bytes, err := json.Marshal(addressList); err != nil {
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

func NewDiscoveryService() *discoveryService {
	return &discoveryService{
		subs: make(map[string][]string),
	}
}
