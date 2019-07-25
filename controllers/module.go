package controllers

import (
	"encoding/json"
	"fmt"
	st "isp-config-service/entity"
	"isp-config-service/model"
	"isp-config-service/module"
	"isp-config-service/socket"

	"isp-config-service/domain"

	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
)

func GetModules(identities []int32) ([]st.Module, error) {
	modules, err := model.ModulesRep.GetModules(identities)
	if err != nil {
		return nil, createUnknownError(err)
	}

	return modules, nil
}

func GetSchemaByModuleId(moduleId int32) (*st.ConfigSchema, error) {
	return model.SchemaRep.GetSchemaByModuleId(moduleId)
}

func GetModulesAggregatedInfo(i structure.Isolation) ([]*domain.ModuleInfo, error) {
	instanceId, err := i.GetInstanceId()
	if err != nil {
		return nil, err
	}

	return module.GetAggregatedModuleInfo(instanceId)
}

func GetActiveModules() ([]st.Module, error) {
	modules, err := model.ModulesRep.GetActiveModules()
	if err != nil {
		return nil, createUnknownError(err)
	}

	return modules, nil
}

func GetConnectedModules() (map[string]interface{}, error) {
	modules := socket.GetRoomsCount()
	bytes, err := json.Marshal(modules)
	if err != nil {
		return nil, createUnknownError(err)
	}
	var response map[string]interface{}
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return response, nil
}

func CreateUpdateModule(module st.Module) (*st.Module, error) {
	if module.Id == 0 {
		instanceExists, err := model.InstanceRep.GetInstanceById(module.InstanceId)
		if err != nil {
			return nil, createUnknownError(err)
		}
		if instanceExists.Id == 0 {
			validationErrors := map[string]string{
				"instanceId": fmt.Sprintf("Entity with instanceId: %d not found", module.InstanceId),
			}
			return nil, utils.CreateValidationErrorDetails(codes.NotFound, utils.ValidationError, validationErrors)
		}
		moduleExists, err := model.ModulesRep.GetModuleByInstanceIdAndName(module.InstanceId, module.Name)
		if err != nil {
			return nil, createUnknownError(err)
		}
		if moduleExists.Id != 0 {
			validationErrors := map[string]string{
				"name": fmt.Sprintf("Entity with name: %s already exists", module.Name),
			}
			return nil, utils.CreateValidationErrorDetails(codes.AlreadyExists, utils.ValidationError, validationErrors)
		}
		module, err = model.ModulesRep.CreateModule(module)
	} else {
		moduleExists, err := model.ModulesRep.GetModuleById(module.Id)
		if err != nil {
			return nil, createUnknownError(err)
		}
		if moduleExists.Id == 0 {
			validationErrors := map[string]string{
				"id": fmt.Sprintf("Entity with id: %d not found", module.Id),
			}
			return nil, utils.CreateValidationErrorDetails(codes.NotFound, utils.ValidationError, validationErrors)
		}
		module, err = model.ModulesRep.UpdateModule(module)
		module.CreatedAt = moduleExists.CreatedAt
		module.Active = moduleExists.Active
	}

	return &module, nil
}

func DeleteModule(identities []int32) (*domain.DeleteResponse, error) {
	if len(identities) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument, utils.ValidationError, validationErrors)
	}
	count, err := model.ModulesRep.DeleteModule(identities)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return &domain.DeleteResponse{Deleted: count}, err
}
