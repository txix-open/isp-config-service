package controller

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/store"
	"isp-config-service/store/state"
)

var Schema *schema

type schema struct {
	rstore *store.Store
}

// GetByModuleId godoc
// @Summary Метод получения схемы конфигурации модуля
// @Description Возвращает текущую json схему конфигурации модуля
// @Tags Схема
// @Accept  json
// @Produce  json
// @Param body body domain.GetByModuleIdRequest true "идентификатор модуля"
// @Success 200 {object} entity.ConfigSchema
// @Failure 404 {object} structure.GrpcError "если схема для модуля не найдена"
// @Failure 500 {object} structure.GrpcError
// @Router /schema/get_by_module_id [POST]
func (c *schema) GetByModuleId(request domain.GetByModuleIdRequest) (*entity.ConfigSchema, error) {
	var result []entity.ConfigSchema
	c.rstore.VisitReadonlyState(func(state state.ReadonlyState) {
		result = state.Schemas().GetByModuleIds([]string{request.ModuleId})
	})
	if len(result) == 0 {
		return nil, status.Errorf(codes.NotFound, "schema with moduleId %s not found", request.ModuleId)
	}
	return &result[0], nil
}

func NewSchema(rstore *store.Store) *schema {
	return &schema{
		rstore: rstore,
	}
}
