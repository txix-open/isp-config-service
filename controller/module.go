package controller

import (
	"google.golang.org/grpc/status"
	"isp-config-service/cluster"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"

	"github.com/integration-system/isp-lib/v2/utils"
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

// GetModuleByName godoc
// @Summary Метод получения модуля по имени
// @Description Модуль
// @Tags Модули
// @Accept json
// @Produce json
// @Param body body domain.GetByModuleNameRequest true "название модуля"
// @Success 200 {object} entity.Module
// @Failure 500 {object} structure.GrpcError
// @Failure 404 {object} structure.GrpcError "модуль с указанным названием не найден"
// @Router /module/get_by_name [POST]
func (c *module) GetModuleByName(req domain.GetByModuleNameRequest) (*entity.Module, error) {
	var module *entity.Module
	c.rstore.VisitReadonlyState(func(readonlyState state.ReadonlyState) {
		module = readonlyState.Modules().GetByName(req.ModuleName)
	})
	if module == nil {
		return nil, status.Errorf(codes.NotFound, "module with name '%s' not found", req.ModuleName)
	}
	return module, nil
}

// @Summary Метод отправки события всем подключенным модулям
// @Description Модуль
// @Tags Модули
// @Accept json
// @Produce json
// @Param body body cluster.BroadcastEvent true "событие"
// @Success 200 ""
// @Failure 500 {object} structure.GrpcError
// @Router /module/broadcast_event [POST]
func (c *module) BroadcastEvent(req cluster.BroadcastEvent) error {
	command := cluster.PrepareBroadcastEventCommand(req)
	err := PerformSyncApply(command, "BroadcastEventCommand", nil)
	if err != nil {
		return err
	}

	return nil
}

func NewModule(rstore *store.Store) *module {
	return &module{
		rstore: rstore,
	}
}
