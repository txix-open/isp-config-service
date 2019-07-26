package controllers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/model"
)

var Schema = schema{}

type schema struct{}

func (schema) GetByModuleId(request domain.GetByModuleIdRequest) (*entity.ConfigSchema, error) {
	if response, err := model.SchemaRep.GetSchemaByModuleId(request.ModuleId); err != nil {
		return nil, err
	} else if response == nil {
		return nil, status.Errorf(codes.NotFound, "schema with moduleId %d not found", request.ModuleId)
	} else {
		return response, nil
	}
}
