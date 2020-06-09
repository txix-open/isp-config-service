package service

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	ConfigHistory = configHistoryService{}
)

type configHistoryService struct{}

func (configHistoryService) HandleDeleteVersionConfigCommand(cfg cluster.Identity,
	state state.WritableState) cluster.ResponseWithError {
	state.WriteableVersionConfigStore().Delete(cfg.Id)
	var (
		deleted int
		err     error
	)
	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		deleted, err = model.VersionStoreRep.Delete(cfg.Id)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"versionConfigId": cfg.Id,
			}).Errorf(codes.DatabaseOperationError, "delete configs: %v", err)
		}
	}
	return cluster.NewResponse(domain.DeleteResponse{Deleted: deleted})
}

func (configHistoryService) HandleGetAllVersionConfigCommand(cfg cluster.Identity,
	state state.WritableState) cluster.ResponseWithError {
	resp := state.VersionConfig().GetByConfigId(cfg.Id)
	return cluster.NewResponse(resp)
}

func (s configHistoryService) SaveConfigVersion(oldConfig entity.Config, writableState state.WritableState) entity.VersionConfig {
	cfg := entity.VersionConfig{
		Id:            state.GenerateId(), //todo all node generate own id
		ConfigVersion: oldConfig.Version,
		ConfigId:      oldConfig.Id,
		Data:          oldConfig.Data,
	}
	removedVersionId := writableState.WriteableVersionConfigStore().Update(cfg)
	if holder.ClusterClient.IsLeader() {
		s.updateDB(cfg, removedVersionId)
	}
	return cfg
}

func (configHistoryService) updateDB(cfg entity.VersionConfig, removedId string) {
	_, err := model.VersionStoreRep.Upsert(cfg)
	if err != nil {
		log.WithMetadata(map[string]interface{}{
			"version_config": cfg,
		}).Errorf(codes.DatabaseOperationError, "upsert version config: %v", err)
	}
	if removedId != "" {
		_, err := model.VersionStoreRep.Delete(removedId)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"version_config_id": removedId,
			}).Errorf(codes.DatabaseOperationError, "delete version config: %v", err)
		}
	}
}
