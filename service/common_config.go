package service

import (
	log "github.com/integration-system/isp-log"
	codes2 "google.golang.org/grpc/codes"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/domain"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	CommonConfigService = commonConfigService{}
)

type commonConfigService struct{}

func (commonConfigService) HandleDeleteConfigsCommand(deleteConfigs cluster.DeleteConfigs, state state.WritableState) cluster.ResponseWithError {
	ids := deleteConfigs.Ids
	affectedConfigs := state.Configs().FilterByCommonConfigs(ids)
	if len(affectedConfigs) > 0 {
		// TODO обсудить формат и сделать ошибку содержательнее
		return cluster.NewResponseErrorf(codes2.InvalidArgument, "cant delete, common configs are in use")
	}
	deleted := state.WritableCommonConfigs().DeleteByIds(ids)
	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.ConfigRep.Delete(ids)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"configIds": ids,
			}).Errorf(codes.DatabaseOperationError, "delete configs: %v", err)
		}
	}
	return cluster.NewResponse(domain.DeleteResponse{Deleted: deleted})
}

func (commonConfigService) HandleUpsertConfigCommand(upsertConfig cluster.UpsertCommonConfig, state state.WritableState) cluster.ResponseWithError {
	config := upsertConfig.Config
	if upsertConfig.Create {
		config = state.WritableCommonConfigs().Create(config)
	} else {
		// Update
		configs := state.CommonConfigs().GetByIds([]string{config.Id})
		if len(configs) == 0 {
			return cluster.NewResponseErrorf(codes2.NotFound, "common config with id %s not found", config.Id)
		}
		config.CreatedAt = configs[0].CreatedAt
		state.WritableCommonConfigs().UpdateById(config)
	}

	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.CommonConfigRep.Upsert(config)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"config": config,
			}).Errorf(codes.DatabaseOperationError, "upsert common config: %v", err)
		}
	}
	if !upsertConfig.Create {
		configsToBroadcast := state.Configs().FilterByCommonConfigs([]string{config.Id})
		ConfigService.BroadcastNewConfig(state, configsToBroadcast...)
	}
	return cluster.NewResponse(config)
}
