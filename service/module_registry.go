package service

import (
	"github.com/integration-system/isp-lib/structure"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
	"sort"
)

var (
	ModuleRegistryService = moduleRegistryService{}
)

type moduleRegistryService struct{}

func (s moduleRegistryService) HandleModuleConnectedCommand(module entity.Module, state state.WritableState) {
	existedModule := state.WritableModules().GetByName(module.Name)
	if existedModule == nil {
		state.WritableModules().Create(module)
	} else {
		module.Id = existedModule.Id
		module.CreatedAt = existedModule.CreatedAt
		module.LastDisconnectedAt = existedModule.LastDisconnectedAt
		state.WritableModules().UpdateByName(module)
	}
	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.ModuleRep.Upsert(module)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"module": module,
			}).Errorf(codes.DatabaseOperationError, "upsert module: %v", err)
		}
	}
}

func (moduleRegistryService) HandleModuleDisconnectedCommand(module entity.Module, state state.WritableState) {
	existedModule := state.WritableModules().GetByName(module.Name)
	if existedModule == nil {
		state.WritableModules().Create(module)
	} else {
		module.Id = existedModule.Id
		module.CreatedAt = existedModule.CreatedAt
		module.LastConnectedAt = existedModule.LastConnectedAt
		state.WritableModules().UpdateByName(module)
	}
	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.ModuleRep.Upsert(module)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"module": module,
			}).Errorf(codes.DatabaseOperationError, "upsert module: %v", err)
		}
	}
}

func (moduleRegistryService) HandleDeleteModulesCommand(deleteModules cluster.DeleteModules, state state.WritableState) int {
	ids := deleteModules.Ids
	deletedModules := state.WritableModules().DeleteByIds(ids)

	configsToDelete := state.WritableConfigs().GetByModuleIds(ids)
	confIds := make([]string, 0, len(configsToDelete))
	for _, conf := range configsToDelete {
		confIds = append(confIds, conf.Id)
	}
	state.WritableConfigs().DeleteByIds(confIds)

	schemasToDelete := state.WritableSchemas().GetByModuleIds(ids)
	schemaIds := make([]string, 0, len(schemasToDelete))
	for _, schema := range schemasToDelete {
		schemaIds = append(schemaIds, schema.Id)
	}
	state.WritableSchemas().DeleteByIds(schemaIds)

	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.ModuleRep.Delete(ids)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"moduleIds": ids,
			}).Errorf(codes.DatabaseOperationError, "delete modules: %v", err)
		}
		_, err = model.ConfigRep.Delete(confIds)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"configIds": confIds,
			}).Errorf(codes.DatabaseOperationError, "delete configs: %v", err)
		}
		_, err = model.SchemaRep.Delete(schemaIds)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"schemaIds": schemaIds,
			}).Errorf(codes.DatabaseOperationError, "delete schemas: %v", err)
		}
	}
	return len(deletedModules)
}

func (moduleRegistryService) GetAggregatedModuleInfo(state state.ReadonlyState) []domain.ModuleInfo {
	modules := state.Modules().GetAll()
	modulesLen := len(modules)
	result := make([]domain.ModuleInfo, 0, modulesLen)
	resMap := make(map[string]domain.ModuleInfo, modulesLen)
	idList := make([]string, modulesLen)
	for i, module := range modules {
		idList[i] = module.Id
		active := state.Mesh().BackendExist(structure.BackendDeclaration{ModuleName: module.Name})
		info := domain.ModuleInfo{
			Id:                 module.Id,
			Name:               module.Name,
			Active:             active,
			CreatedAt:          module.CreatedAt,
			LastConnectedAt:    module.LastConnectedAt,
			LastDisconnectedAt: module.LastDisconnectedAt,
		}
		resMap[module.Id] = info
	}

	configs := state.Configs().GetByModuleIds(idList)
	for _, cfg := range configs {
		info := resMap[cfg.ModuleId]
		info.Configs = append(info.Configs, cfg)
		resMap[cfg.ModuleId] = info
	}
	schemas := state.Schemas().GetByModuleIds(idList)
	for _, s := range schemas {
		info := resMap[s.ModuleId]
		schema := s.Schema
		info.ConfigSchema = &schema
		resMap[s.ModuleId] = info
	}

	for key := range resMap {
		info := resMap[key]
		backends := state.Mesh().GetBackends(info.Name)
		conns := make([]domain.Connection, 0, len(backends))

		for _, back := range backends {
			con := domain.Connection{
				Version:    back.Version,
				LibVersion: back.LibVersion,
				Address:    back.Address,
				Endpoints:  back.Endpoints,
			}
			sort.Slice(con.Endpoints, func(i, j int) bool {
				return con.Endpoints[i].Path < con.Endpoints[j].Path
			})
			con.EstablishedAt = info.LastConnectedAt
			conns = append(conns, con)
		}
		sort.Slice(conns, func(i, j int) bool {
			return conns[i].EstablishedAt.Before(conns[j].EstablishedAt)
		})
		info.Status = conns
		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
