package controllers

import (
	"fmt"

	st "isp-config-service/entity"
	"isp-config-service/model"
	"isp-config-service/socket"

	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-config-service/domain"
)

func GetConfigs(identities []int64) ([]st.Config, error) {
	configs, err := model.ConfigRep.GetConfigs(identities)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return configs, nil
}

func GetConfigByInstanceUUIDAndModuleName(request st.ModuleInstanceIdentity) (*st.Config, error) {
	config, err := model.ConfigRep.GetConfigByInstanceUUIDAndModuleName(request.Uuid, request.Name)
	if err != nil {
		return nil, createUnknownError(err)
	}
	if config == nil {
		return nil, status.Error(codes.NotFound, utils.ValidationError)
	}

	return config, nil
}

func CreateUpdateConfig(config st.Config) (*st.Config, error) {
	var err error
	var cfg *st.Config
	if config.Id == 0 {
		cfg, err = model.ConfigRep.CreateConfig(&config)
		if err != nil {
			return nil, err
		}
	} else {
		configExists, err := model.ConfigRep.GetConfigById(config.Id)
		if err != nil {
			return nil, err
		}
		if configExists == nil {
			validationErrors := map[string]string{
				"id": fmt.Sprintf("Entity with id: %d not found", config.Id),
			}
			return nil, utils.CreateValidationErrorDetails(codes.NotFound, utils.ValidationError, validationErrors)
		}
		cfg, err = model.ConfigRep.UpdateConfig(&config)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Active {
		moduleInstanceIdentity, err := model.ConfigRep.GetModuleNameAndInstanceUUIDByConfigId(cfg.Id)
		if err != nil {
			return nil, createUnknownError(err)
		}
		socket.SendNewConfig(moduleInstanceIdentity.Uuid, moduleInstanceIdentity.Name, cfg)
	}

	return cfg, nil
}

func UpdateActiveConfigByInstanceUUIDAndModuleName(configReq domain.ConfigInstanceModuleName) (*st.Config, error) {
	config, err := model.ConfigRep.GetConfigByInstanceUUIDAndModuleName(configReq.InstanceUuid, configReq.ModuleName)
	if err != nil || config == nil {
		validationErrors := map[string]string{
			"id": fmt.Sprintf("Active config for module: %s and instance uuid: %s not found",
				configReq.ModuleName, configReq.InstanceUuid),
		}
		return nil, utils.CreateValidationErrorDetails(codes.NotFound, utils.ValidationError, validationErrors)
	}
	config.Data = configReq.ConfigData
	return CreateUpdateConfig(*config)
}

func MarkConfigAsActive(identity domain.LongIdentitiesRequest) (*st.Config, error) {
	configExists, err := model.ConfigRep.GetConfigById(identity.Id)
	if err != nil {
		return nil, createUnknownError(err)
	}
	if configExists == nil {
		validationErrors := map[string]string{
			"id": fmt.Sprintf("Entity with id: %d not found", identity.Id),
		}
		return nil, utils.CreateValidationErrorDetails(codes.NotFound, utils.ValidationError, validationErrors)
	}
	configExists.Active = true
	configExists, err = model.ConfigRep.UpdateConfig(configExists)
	if err != nil {
		return nil, err
	}

	if configExists.Active {
		moduleInstanceIdentity, err := model.ConfigRep.GetModuleNameAndInstanceUUIDByConfigId(configExists.Id)
		if err != nil {
			return nil, createUnknownError(err)
		}
		socket.SendNewConfig(moduleInstanceIdentity.Uuid, moduleInstanceIdentity.Name, configExists)
	}

	return configExists, nil
}

func DeleteConfig(identities []int64) (*domain.DeleteResponse, error) {
	if len(identities) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument, utils.ValidationError, validationErrors)
	}
	count, err := model.ConfigRep.DeleteConfig(identities)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return &domain.DeleteResponse{Deleted: count}, err
}
