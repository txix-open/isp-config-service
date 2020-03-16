package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/holder"
	"isp-config-service/store/state"
)

const (
	messagesBackoffInterval   = 100 * time.Millisecond
	messagesBackoffMaxRetries = 3
)

var (
	DiscoveryService = NewDiscoveryService()
)

type discoveryService struct {
	subs map[string][]string
	lock sync.RWMutex
}

func (ds *discoveryService) HandleDisconnect(connID string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	if events, ok := ds.subs[connID]; ok {
		holder.EtpServer.Rooms().LeaveByConnId(connID, events...)
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
	events := make([]string, 0, len(modules))
	for _, module := range modules {
		event := utils.ModuleConnected(module)
		addressList := mesh.GetModuleAddresses(module)
		eventsAddresses = append(eventsAddresses, addressList)
		events = append(events, event)
	}
	ds.subs[conn.ID()] = events
	holder.EtpServer.Rooms().Join(conn, events...)

	go func(events []string, eventsAddresses [][]structure.AddressConfiguration, conn etp.Conn) {
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
	go func(room, event string, addressList []structure.AddressConfiguration) {
		err := ds.broadcastAddrList(room, event, addressList)
		if err != nil {
			log.Errorf(codes.DiscoveryServiceSendModulesError, "broadcast module connected %v", err)
		}
	}(event, event, addressList)
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

func NewDiscoveryService() *discoveryService {
	return &discoveryService{
		subs: make(map[string][]string),
	}
}
