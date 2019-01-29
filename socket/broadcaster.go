package socket

import (
	"github.com/googollee/go-socket.io"
	"github.com/integration-system/isp-lib/utils"
	"sync"
	"time"
)

const (
	defaultTimeout = 1 * time.Second
)

type Record struct {
	data   interface{}
	future *time.Timer
}

type broadcaster struct {
	Server                *socketio.Server
	AwaitNewRecordTimeout time.Duration

	configMap    map[string]*Record
	routesMap    map[string]*Record
	converterMap map[string]*Record
	routerMap    map[string]*Record

	configLock    sync.Mutex
	routesLock    sync.Mutex
	converterLock sync.Mutex
	routerLock    sync.Mutex
}

func (b *broadcaster) BroadcastConfig(moduleKey string, data interface{}, callback func(data interface{})) {
	b.doBroadcast(moduleKey, data, utils.ConfigSendConfigChanged, &b.configLock, b.configMap, callback)
}

func (b *broadcaster) BroadcastRoutes(instanceId string, data interface{}, callback func(data interface{})) {
	b.doBroadcast(instanceId, data, utils.ConfigSendRoutesChanged, &b.routesLock, b.routesMap, callback)
}

/*func (b *broadcaster) BroadcastConverterAddressList(instanceId string, data interface{}, callback func(data interface{})) {
	b.doBroadcast(instanceId, data, utils.SendNewConverterConnected, &b.converterLock, b.converterMap, callback)
}

func (b *broadcaster) BroadcastRouterAddressList(instanceId string, data interface{}, callback func(data interface{})) {
	b.doBroadcast(instanceId, data, utils.SendNewRouterConnected, &b.routerLock, b.routerMap, callback)
}*/

func (b *broadcaster) doBroadcast(
	key string,
	data interface{},
	event string,
	lock *sync.Mutex,
	store map[string]*Record,
	callback func(data interface{}),
) {
	lock.Lock()
	defer lock.Unlock()

	if r, present := store[key]; present {
		r.data = data
		if r.future.Stop() {
			r.future.Reset(b.AwaitNewRecordTimeout)
		}
	} else {
		r := &Record{
			data: data,
		}
		r.future = time.AfterFunc(b.AwaitNewRecordTimeout, func() {
			lock.Lock()
			if r.data != nil {
				b.Server.BroadcastTo(key, event, r.data)
				if callback != nil {
					callback(r.data)
				}
			}
			delete(store, key)
			lock.Unlock()
		})

		store[key] = r
	}
}

func NewBroadcaster(server *socketio.Server, awaitNewRecordTimeout time.Duration) *broadcaster {
	timeout := awaitNewRecordTimeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &broadcaster{
		Server:                server,
		AwaitNewRecordTimeout: timeout,
		configMap:             map[string]*Record{},
		routesMap:             map[string]*Record{},
		converterMap:          map[string]*Record{},
		routerMap:             map[string]*Record{},
	}
}
