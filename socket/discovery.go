package socket

import (
	"encoding/json"
	"github.com/cenkalti/backoff"
	"github.com/googollee/go-socket.io"
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"sync"
	"time"
)

var (
	discoverersMap  = make(map[string]*Discoverer)
	discoverersLock sync.RWMutex
)

type channelIdList []string

type Discoverer struct {
	subs       map[string]channelIdList
	instanceId string
	lock       sync.RWMutex
}

func (d *Discoverer) OnDisconnection(channelId string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	newSubs := make(map[string]channelIdList, len(d.subs))

	for event, list := range d.subs {
		newList := make(channelIdList, 0)
		for _, id := range list {
			if id != channelId {
				newList = append(newList, id)
			}
		}
		if len(newList) > 0 {
			newSubs[event] = newList
		}
	}

	d.subs = newSubs
}

func (d *Discoverer) Subscribe(events []string, socket socketio.Socket) {
	if len(events) == 0 {
		return
	}
	channelId := socket.Id()

	d.lock.Lock()
	defer d.lock.Unlock()

	connectedModulesEvents := getConnectedModulesEventMap(d.instanceId)

	for _, event := range events {
		if list, ok := d.subs[event]; ok {
			d.subs[event] = append(list, channelId)
		} else {
			d.subs[event] = []string{channelId}
		}

		if addrList, ok := connectedModulesEvents[event]; ok {
			d.sendAddrList(socket, event, addrList)
		}
	}
}

func (d *Discoverer) BroadcastModuleAddresses(moduleName string) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	event := utils.ModuleConnected(moduleName)
	connectedModulesEvents := getConnectedModulesEventMap(d.instanceId)
	addrList := connectedModulesEvents[event]
	channelList, channelListPresent := d.subs[event]
	if channelListPresent {
		for _, id := range channelList {
			if s, ok := GetConnectionById(d.instanceId, id); ok {
				d.sendAddrList(s, event, addrList)
			}
		}
	}
}

func (d *Discoverer) sendAddrList(socket socketio.Socket, event string, addressList []structure.AddressConfiguration) {
	if addressList == nil {
		addressList = []structure.AddressConfiguration{}
	}
	if bytes, err := json.Marshal(addressList); err != nil {
		logger.Warn(err)
	} else {
		bf := backoff.WithMaxRetries(backoff.NewConstantBackOff(100*time.Millisecond), 3)
		err := backoff.Retry(func() error {
			return socket.Emit(event, string(bytes))
		}, bf)
		if err != nil {
			logger.Error(err)
		}
	}
}

func GetModuleAddressList(instanceId string, moduleName string) []structure.AddressConfiguration {
	event := utils.ModuleConnected(moduleName)
	connectedModulesEvents := getConnectedModulesEventMap(instanceId)
	addrList := connectedModulesEvents[event]
	if addrList == nil {
		return make([]structure.AddressConfiguration, 0)
	}
	return addrList
}

func getConnectedModulesEventMap(instanceId string) map[string][]structure.AddressConfiguration {
	connectedModulesEventMap := make(map[string][]structure.AddressConfiguration)
	connMap := GetModuleConnectionIdMapByInstanceId(instanceId)

	for moduleName, channelIdList := range connMap {
		event := utils.ModuleConnected(moduleName)
		addrList := make([]structure.AddressConfiguration, 0)

		backendConfigs := GetRoutes().Routes
		for _, id := range channelIdList {
			if cfg, ok := backendConfigs[id]; ok {
				addrList = append(addrList, cfg.Address)
			}
		}
		connectedModulesEventMap[event] = addrList
	}
	return connectedModulesEventMap
}

func getOrRegisterDiscoverer(instanceId string) *Discoverer {
	discoverersLock.RLock()
	d, ok := discoverersMap[instanceId]
	discoverersLock.RUnlock()
	if ok {
		return d
	}

	discoverersLock.Lock()
	defer discoverersLock.Unlock()
	if d, ok := discoverersMap[instanceId]; ok {
		return d
	}
	d = &Discoverer{
		subs:       make(map[string]channelIdList),
		instanceId: instanceId,
	}
	discoverersMap[instanceId] = d
	return d
}
