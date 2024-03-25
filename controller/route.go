package controller

import (
	"github.com/integration-system/isp-lib/v2/structure"
	"isp-config-service/store"
	"isp-config-service/store/state"
)

var (
	Routes *routes
)

type routes struct {
	rstore *store.Store
}

// @Summary Метод получения маршрутов
// @Description Возвращает все доступные модули
// @Tags Роуты
// @Accept  json
// @Produce  json
// @Success 200 {array} structure.BackendDeclaration
// @Router /routing/get_routes [POST]
func (c *routes) GetRoutes() ([]structure.BackendDeclaration, error) {
	var response structure.RoutingConfig
	c.rstore.VisitReadonlyState(func(state state.ReadonlyState) {
		response = state.Mesh().GetRoutes()
	})
	return response, nil
}

func NewRoutes(rstore *store.Store) *routes {
	return &routes{
		rstore: rstore,
	}
}
