package service

import (
	"encoding/json"
	"fmt"

	"github.com/integration-system/bellows"
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	codes2 "google.golang.org/grpc/codes"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	ConfigService = configService{}
)

type configService struct{}

type validationSchemaError struct {
	Description map[string]string
}

func (e validationSchemaError) Error() string {
	resp := ""
	for field, desc := range e.Description {
		resp = fmt.Sprintf("%s- %s%s\n", resp, field, desc)
	}
	return resp
}

func (cs configService) GetCompiledConfig(moduleName string, state state.ReadonlyState) (map[string]interface{}, error) {
	module := state.Modules().GetByName(moduleName)
	if module == nil {
		return nil, errors.Errorf("module with name %s not found", moduleName)
	}
	config := state.Configs().GetActiveByModuleId(module.Id)
	if config == nil {
		return nil, errors.Errorf("no active configs for moduleName %s, moduleId %s", moduleName, module.Id)
	}

	return cs.CompileConfig(config.Data, state, config.CommonConfigs...), nil
}

func (configService) CompileConfig(
	data map[string]interface{}, state state.ReadonlyState, commonConfigsIds ...string) map[string]interface{} {
	commonConfigs := state.CommonConfigs().GetByIds(commonConfigsIds)
	configsToMerge := make([]map[string]interface{}, 0, len(commonConfigs))
	for _, common := range commonConfigs {
		configsToMerge = append(configsToMerge, common.Data)
	}
	configsToMerge = append(configsToMerge, data)

	return mergeNestedMaps(configsToMerge...)
}

func (cs configService) HandleActivateConfigCommand(
	activateConfig cluster.ActivateConfig, state state.WritableState) cluster.ResponseWithError {
	configs := state.Configs().GetByIds([]string{activateConfig.ConfigId})
	if len(configs) == 0 {
		return cluster.NewResponseErrorf(codes2.NotFound, "config with id %s not found", activateConfig.ConfigId)
	}
	config := configs[0]
	affected := state.WritableConfigs().Activate(config, activateConfig.Date)
	if holder.ClusterClient.IsLeader() {
		for i := range affected {
			// TODO handle db errors
			_, err := model.ConfigRep.Upsert(affected[i])
			if err != nil {
				log.WithMetadata(map[string]interface{}{
					"config": config,
				}).Errorf(codes.DatabaseOperationError, "upsert config: %v", err)
			}
		}
	}
	config = affected[len(affected)-1]
	cs.BroadcastActiveConfigs(state, config)
	return cluster.NewResponse(config)
}

func (configService) HandleDeleteConfigsCommand(deleteConfigs cluster.DeleteConfigs, state state.WritableState) int {
	ids := deleteConfigs.Ids
	deleted := state.WritableConfigs().DeleteByIds(ids)
	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.ConfigRep.Delete(ids)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"configIds": ids,
			}).Errorf(codes.DatabaseOperationError, "delete configs: %v", err)
		}
	}
	return deleted
}

func (cs configService) HandleUpsertConfigCommand(upsertConfig cluster.UpsertConfig,
	state state.WritableState) cluster.ResponseWithError {
	config := upsertConfig.Config
	module := state.Modules().GetById(config.ModuleId)
	if module == nil {
		return cluster.NewResponseErrorf(codes2.NotFound, "moduleId %s not found", config.ModuleId)
	}

	schemaStorage := state.Schemas().GetByModuleIds([]string{config.ModuleId})
	if len(schemaStorage) == 0 {
		return cluster.NewResponseErrorf(codes2.NotFound, "schema for moduleId %s not found", config.ModuleId)
	}
	dataForValidate := cs.CompileConfig(config.Data, state, config.CommonConfigs...)
	validSchema, err := cs.validateSchema(schemaStorage[0], dataForValidate)

	if !upsertConfig.Unsafe && err != nil {
		switch err := err.(type) {
		case validationSchemaError:
			return cluster.NewResponse(
				domain.CreateUpdateConfigResponse{
					Config:       nil,
					ErrorDetails: err.Description,
				})
		default:
			return cluster.NewResponseErrorf(codes2.Internal, "%v", err)
		}
	}

	if upsertConfig.Create {
		config = state.WritableConfigs().Create(config)
	} else {
		// Update
		configs := state.Configs().GetByIds([]string{config.Id})
		if len(configs) == 0 {
			return cluster.NewResponseErrorf(codes2.NotFound, "config with id %s not found", config.Id)
		}
		oldCfg := configs[0]
		config.CreatedAt = oldCfg.CreatedAt
		config.Active = oldCfg.Active
		config.Version = oldCfg.Version + 1
		state.WritableConfigs().UpdateById(config)

		ConfigHistory.SaveConfigVersion(upsertConfig.VersionId, oldCfg, state, upsertConfig.VersionCreateAt)
	}

	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.ConfigRep.Upsert(config)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"config": config,
			}).Errorf(codes.DatabaseOperationError, "upsert config: %v", err)
		}
	}

	cs.BroadcastActiveConfigs(state, config)
	return cluster.NewResponse(domain.CreateUpdateConfigResponse{
		Config: &domain.ConfigModuleInfo{
			Config: config,
			Valid:  validSchema,
		},
		ErrorDetails: nil,
	})
}

func (cs configService) BroadcastActiveConfigs(state state.ReadonlyState, configs ...entity.Config) {
	for i := range configs {
		cfg := &configs[i]
		if !cfg.Active {
			continue
		}
		moduleID := cfg.ModuleId
		module := state.Modules().GetById(moduleID)
		moduleName := module.Name
		room := Room.Module(moduleName)

		compiledConfig, err := cs.GetCompiledConfig(moduleName, state)
		if err != nil {
			go cs.broadcast(room, utils.ConfigError, []byte(err.Error()))
			continue
		}
		data, err := json.Marshal(compiledConfig)
		if err != nil {
			go cs.broadcast(room, utils.ConfigError, []byte(err.Error()))
			continue
		}
		go cs.broadcast(room, utils.ConfigSendConfigChanged, data)
	}
}

func (cs configService) broadcast(room, event string, body []byte) {
	conns := holder.EtpServer.Rooms().ToBroadcast(room)
	for _, conn := range conns {
		err := EmitConnWithTimeout(conn, event, body)
		if err != nil {
			log.Errorf(codes.ConfigServiceBroadcastConfigError, "broadcast '%s' to %s: %v", event, conn.RemoteAddr(), err)
		}
	}
}

func (configService) validateSchema(schema entity.ConfigSchema, data map[string]interface{}) (bool, error) {
	schemaLoader := gojsonschema.NewGoLoader(schema.Schema)
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, err
	} else if result.Valid() {
		return true, nil
	}

	desc := make(map[string]string)
	for _, value := range result.Errors() {
		desc[value.Field()] = value.Description()
	}
	return false, validationSchemaError{Description: desc}
}

func mergeNestedMaps(maps ...map[string]interface{}) map[string]interface{} {
	if len(maps) == 1 {
		return maps[0]
	}
	result := bellows.Flatten(maps[0])
	for i := 1; i < len(maps); i++ {
		newFlatten := bellows.Flatten(maps[i])
		for k, v := range newFlatten {
			result[k] = v
		}
	}
	return bellows.Expand(result).(map[string]interface{})
}
