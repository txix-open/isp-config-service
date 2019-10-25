package controller

import (
	"isp-config-service/cluster"
	"isp-config-service/domain"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"

	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
)

var Module *module

type module struct {
	rstore *store.Store
}

// GetModulesAggregatedInfo godoc
// @Summary Метод получения полной информации о состоянии модуля
// @Description Возвращает полное состояние всех модулей в кластере (конфигурация, схема конфигурации, подключенные экземпляры)
// @Tags Модули
// @Accept  json
// @Produce  json
// @Success 200 {array} domain.ModuleInfo
// @Failure 500 {object} structure.GrpcError
// @Router /module/get_modules_info [POST]
func (c *module) GetModulesAggregatedInfo() ([]domain.ModuleInfo, error) {
	var response []domain.ModuleInfo
	c.rstore.VisitReadonlyState(func(state state.ReadonlyState) {
		response = service.ModuleRegistryService.GetAggregatedModuleInfo(state)
	})
	return response, nil
}

// DeleteModules godoc
// @Summary Метод удаления объектов модулей по идентификаторам
// @Description Возвращает количество удаленных модулей
// @Tags Модули
// @Accept  json
// @Produce  json
// @Param body body []string true "массив идентификаторов модулей"
// @Success 200 {object} domain.DeleteResponse
// @Failure 500 {object} structure.GrpcError
// @Router /module/delete_module [POST]
func (*module) DeleteModules(identities []string) (*domain.DeleteResponse, error) {
	if len(identities) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument, utils.ValidationError, validationErrors)
	}

	var deleteResponse domain.DeleteResponse
	command := cluster.PrepareDeleteModulesCommand(identities)
	err := PerformSyncApply(command, "DeleteCommonConfigsCommand", &deleteResponse)
	if err != nil {
		return nil, err
	}
	return &deleteResponse, err
}

func NewModule(rstore *store.Store) *module {
	return &module{
		rstore: rstore,
	}
}
