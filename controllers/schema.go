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

// GetModulesAggregatedInfo godoc
// @Summary Метод схемы конфигурации модуля
// @Description Возвращает текущую json схему конфигурации модуля
// @Accept  json
// @Produce  json
// @Param body body domain.GetByModuleIdRequest true "идентификатор модуля"
// @Success 200 {array} entity.ConfigSchema
// @Failure 404 {object} structure.GrpcError "если схема для модуля не найдена"
// @Failure 500 {object} structure.GrpcError
// @Router /config/get_modules_info [POST]
func (schema) GetByModuleId(request domain.GetByModuleIdRequest) (*entity.ConfigSchema, error) {
	if response, err := model.SchemaRep.GetSchemaByModuleId(request.ModuleId); err != nil {
		return nil, err
	} else if response == nil {
		return nil, status.Errorf(codes.NotFound, "schema with moduleId %d not found", request.ModuleId)
	} else {
		return response, nil
	}
}
