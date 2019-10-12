package service

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
	"time"
)

var (
	ModuleRegistryService = moduleRegistryService{}
)

type moduleRegistryService struct{}

func (s moduleRegistryService) HandleModuleConnectedCommand(moduleConnected cluster.ModuleConnected, state state.WritableState) error {
	module := s.updateModuleLastConnected(moduleConnected.ModuleName, state.WritableModules())
	if holder.ClusterClient.IsLeader() {
		_, err := model.ModuleRep.Upsert(module)
		if err != nil {
			log.Errorf(codes.DatabaseOperationError, "upsert module: %v", err)
		}
	}
	return nil
}

func (moduleRegistryService) updateModuleLastConnected(moduleName string, moduleStore state.WriteableModuleStore) entity.Module {
	existedModule := moduleStore.GetByName(moduleName)
	if existedModule == nil {
		module := moduleStore.Create(moduleName)
		return module
	}
	existedModule.LastConnectedAt = time.Now()
	moduleStore.Update(*existedModule)
	return *existedModule
}

func (moduleRegistryService) updateModuleLastDisconnected(moduleName string, moduleStore state.WriteableModuleStore) entity.Module {
	existedModule := moduleStore.GetByName(moduleName)
	if existedModule == nil {
		module := moduleStore.Create(moduleName)
		return module
	}
	existedModule.LastDisconnectedAt = time.Now()
	moduleStore.Update(*existedModule)
	return *existedModule
}
