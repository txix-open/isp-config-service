package controllers

import (
	"fmt"

	st "isp-config-service/entity"
	"isp-config-service/model"

	"isp-config-service/domain"

	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetInstances(identities []int32) ([]st.Instance, error) {
	instances, err := model.InstanceRep.GetInstances(identities)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return instances, err
}

func CreateUpdateInstance(instance st.Instance) (*st.Instance, error) {
	if instance.Id == 0 {
		moduleExists, err := model.InstanceRep.GetInstanceByUuid(instance.Uuid)
		if err != nil {
			return nil, createUnknownError(err)
		}
		if moduleExists.Id != 0 {
			validationErrors := map[string]string{
				"name": fmt.Sprintf("Entity with uuid: %s already exists", instance.Name),
			}
			return nil, utils.CreateValidationErrorDetails(codes.AlreadyExists, utils.ValidationError, validationErrors)
		}
		instance, err = model.InstanceRep.CreateInstance(instance)
	} else {
		instanceExists, err := model.InstanceRep.GetInstanceById(instance.Id)
		if err != nil {
			return nil, createUnknownError(err)
		}
		if instanceExists.Id == 0 {
			validationErrors := map[string]string{
				"id": fmt.Sprintf("Entity with id: %d not found", instance.Id),
			}
			return nil, utils.CreateValidationErrorDetails(codes.NotFound, utils.ValidationError, validationErrors)
		}
		instance, err = model.InstanceRep.UpdateInstance(instance)
		instance.CreatedAt = instanceExists.CreatedAt
	}

	return &instance, nil
}

func DeleteInstance(identities []int32) (*domain.DeleteResponse, error) {
	if len(identities) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument, utils.ValidationError, validationErrors)
	}
	count, err := model.InstanceRep.DeleteInstance(identities)
	if err != nil {
		return nil, createUnknownError(err)
	}
	return &domain.DeleteResponse{Deleted: count}, err
}

func createUnknownError(err error) error {
	logger.Error(err)
	return status.Error(codes.Unknown, utils.ServiceError)
}
