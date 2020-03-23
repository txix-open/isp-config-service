package controller

import (
	"github.com/integration-system/isp-lib/v2/utils"
	log "github.com/integration-system/isp-log"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-config-service/cluster"
	codes2 "isp-config-service/codes"
	"isp-config-service/holder"
)

var json = jsoniter.ConfigFastest

func unknownError() error {
	return status.Error(codes.Unknown, utils.ServiceError)
}

func PerformSyncApply(command []byte, responsePtr interface{}) error {
	applyLogResponse, err := holder.ClusterClient.SyncApply(command)
	if err != nil {
		cmd := cluster.ParseCommand(command)
		log.Warnf(codes2.SyncApplyError, "apply %s: %v", cmd, err)
		return unknownError()
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		commandName := cluster.ParseCommand(command).String()
		log.WithMetadata(map[string]interface{}{
			"result":      string(applyLogResponse.Result),
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes2.SyncApplyError, "apply command")
		return unknownError()
	}
	if responsePtr != nil {
		err = json.Unmarshal(applyLogResponse.Result, responsePtr)
		if err != nil {
			return unknownError()
		}
	}
	return nil
}

func PerformSyncApplyWithError(command []byte, responsePtr interface{}) error {
	applyLogResponse, err := holder.ClusterClient.SyncApply(command)
	if err != nil {
		cmd := cluster.ParseCommand(command)
		log.Warnf(codes2.SyncApplyError, "apply %s: %v", cmd, err)
		return unknownError()
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		commandName := cluster.ParseCommand(command).String()
		log.WithMetadata(map[string]interface{}{
			"result":      string(applyLogResponse.Result),
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes2.SyncApplyError, "apply command")
		return unknownError()
	}
	clusterResp := cluster.ResponseWithError{
		Response: responsePtr,
	}
	err = json.Unmarshal(applyLogResponse.Result, &clusterResp)
	if err != nil {
		return unknownError()
	}
	if clusterResp.Error != "" {
		return status.Error(clusterResp.ErrorCode, clusterResp.Error)
	}
	return nil
}
