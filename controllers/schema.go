package controllers

import (
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/model"
)

var Schema = schema{}

type schema struct{}

func (schema) GetByModuleId(request domain.GetByModuleIdRequest) (*entity.ConfigSchema, error) {
	return model.SchemaRep.GetSchemaByModuleId(request.ModuleId)
}
