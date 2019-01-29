package service

import (
	"sync"

	libSt "github.com/integration-system/isp-lib/structure"
)

type RouterTables struct {
	Routes map[string]libSt.BackendDeclaration
	sync.RWMutex
}

func (routerTables *RouterTables) DeleteRoute(socketId string) bool {
	routerTables.Lock()
	defer routerTables.Unlock()

	changed := false
	for k := range (*routerTables).Routes {
		if k == socketId {
			delete((*routerTables).Routes, k)
			changed = true
		}
	}
	return changed
}

func (routerTables *RouterTables) GetRoutes() *libSt.RoutingConfig {
	routerTables.RLock()
	defer routerTables.RUnlock()

	var routes = libSt.RoutingConfig{}
	for _, v := range (*routerTables).Routes {
		routes = append(routes, v)
	}
	return &routes
}

func (routerTables *RouterTables) AddAddressOrUpdate(socketId string, backendConfig libSt.BackendDeclaration) bool {
	routerTables.Lock()
	defer routerTables.Unlock()

	exists := false
	changed := false
	for k, v := range (*routerTables).Routes {
		if k == socketId {
			if !v.IsPathsEqual(backendConfig.Endpoints) {
				(*routerTables).Routes[socketId] = backendConfig
				changed = true
			}
			exists = true
		}
	}
	if !exists {
		if (*routerTables).Routes != nil {
			(*routerTables).Routes[socketId] = backendConfig
		} else {
			(*routerTables).Routes = map[string]libSt.BackendDeclaration{socketId: backendConfig}
		}
		changed = true
	}
	return changed
}
