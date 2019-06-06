package service

import (
	"errors"
	"fmt"
	"time"

	"isp-config-service/entity"
	"isp-config-service/model"

	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
)

func Disonnect(instanceUuid string, moduleName string) error {
	err := model.ModulesRep.UpdateModuleDisconnect(instanceUuid, moduleName)
	if err != nil {
		logger.Warn(err)
		return errors.New(utils.ServiceError)
	}
	return nil
}

func NewConnection(instanceUuid string, moduleName string) error {
	instance, err := model.InstanceRep.GetInstanceByUuid(instanceUuid)
	if err != nil {
		logger.Error(err)
		return errors.New(utils.ServiceError)
	}
	if instance.Id == 0 {
		return errors.New(fmt.Sprintf("The instance with uuid: %s not found", instanceUuid))
	}
	module, err := model.ModulesRep.GetModuleByInstanceIdAndName(instance.Id, moduleName)
	if err != nil {
		logger.Error(err)
		return errors.New(utils.ServiceError)
	}
	if module.Id == 0 {
		module.Name = moduleName
		module.InstanceId = instance.Id
		module.LastConnectedAt = time.Now()
		module, err = model.ModulesRep.CreateModule(module)
		if err != nil {
			logger.Error(err)
			return errors.New(utils.ServiceError)
		}
		return nil
	} else {
		module.LastConnectedAt = time.Now()
		module, err = model.ModulesRep.UpdateModule(module)
		if err != nil {
			logger.Error(err)
			return errors.New(utils.ServiceError)
		}
		return nil
	}
}

func GetConfig(instanceUuid string, moduleName string) (*entity.Config, error) {
	config, err := model.ConfigRep.GetConfigByInstanceUUIDAndModuleName(instanceUuid, moduleName)
	if err != nil {
		logger.Error(err)
		return nil, errors.New(utils.ServiceError)
	}
	if config == nil {
		errorMessage := "Active config for " + instanceUuid + ":" + moduleName + " not found"
		logger.Warn(errorMessage)
		return nil, errors.New(errorMessage)
	}
	return config, nil
}
