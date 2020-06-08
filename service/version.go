package service

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/domain"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	VersionConfig = versionConfigService{}
)

type versionConfigService struct{}

func (versionConfigService) HandleDeleteVersionConfigCommand(cfg cluster.Identity,
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

func (versionConfigService) HandleGetAllVersionConfigCommand(cfg cluster.Identity,
	state state.WritableState) cluster.ResponseWithError {
	resp := state.VersionConfig().GetByConfigId(cfg.Id)
	return cluster.NewResponse(resp)
}
