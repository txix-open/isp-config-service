package controller

import (
	"isp-config-service/cluster"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/service"
	"isp-config-service/store"
	"isp-config-service/store/state"
)

var CommonConfig *commonConfig

type commonConfig struct {
	rstore *store.Store
}

// @Summary Метод получения объектов конфигурации по идентификаторам
// @Description Возвращает массив конфигураций по запрошенным идентификаторам (все, если массив пустой)
// @Tags Общие конфигурации
// @Accept json
// @Produce json
// @Success 200 {array} entity.CommonConfig
// @Router /common_config/get_configs [POST]
func (c *commonConfig) GetConfigs(identities []string) []entity.CommonConfig {
	var response []entity.CommonConfig
	c.rstore.VisitReadonlyState(func(state state.ReadonlyState) {
		if len(identities) > 0 {
			response = state.CommonConfigs().GetByIds(identities)
		} else {
			response = state.CommonConfigs().GetAll()
		}
	})
	return response
}

// @Summary Метод обновления общей конфигурации
// @Description Если конфиг с таким id существует, то обновляет данные, если нет, то добавляет данные в базу
// @Description В случае обновления рассылает всем подключенным модулям актуальную конфигурацию
// @Tags Общие конфигурации
// @Accept json
// @Produce json
// @Param body body entity.CommonConfig true "объект для сохранения"
// @Success 200 {object} entity.CommonConfig
// @Failure 404 {object} structure.GrpcError "если конфигурация не найдена"
// @Failure 500 {object} structure.GrpcError
// @Router /common_config/create_update_config [POST]
func (c *commonConfig) CreateUpdateConfig(config entity.CommonConfig) (*entity.CommonConfig, error) {
	var response entity.CommonConfig
	now := state.GenerateDate()
	config.CreatedAt = now
	config.UpdatedAt = now
	upsertConfig := cluster.UpsertCommonConfig{
		Config: config,
	}
	if config.Id == "" {
		upsertConfig.Config.Id = state.GenerateId()
		upsertConfig.Create = true
	}
	command := cluster.PrepareUpsertCommonConfigCommand(upsertConfig)
	err := PerformSyncApplyWithError(command, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// @Summary Метод удаления объектов общей конфигурации по идентификаторам
// @Description Возвращает флаг удаления и набор связей с модулями и конфигурациями, в случае наличия связей deleted всегда false
// @Tags Общие конфигурации
// @Accept json
// @Produce json
// @Param body body domain.ConfigIdRequest true "id общей конфигурации"
// @Success 200 {object} domain.DeleteCommonConfigResponse
// @Failure 500 {object} structure.GrpcError
// @Router /common_config/delete_config [POST]
func (c *commonConfig) DeleteConfigs(req domain.ConfigIdRequest) (*domain.DeleteCommonConfigResponse, error) {
	var response domain.DeleteCommonConfigResponse
	command := cluster.PrepareDeleteCommonConfigsCommand(req.Id)
	err := PerformSyncApplyWithError(command, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// @Summary Метод компиляции итоговой конфигурации для модулей
// @Description Возвращает скомпилированный объект конфигурации
// @Tags Общие конфигурации
// @Accept json
// @Produce json
// @Param body body domain.CompileConfigsRequest true "перечисление идентификаторов общей конфигурации и исхдных конфиг"
// @Success 200 {object} domain.CompiledConfigResponse
// @Router /common_config/compile [POST]
func (c *commonConfig) CompileConfigs(req domain.CompileConfigsRequest) domain.CompiledConfigResponse {
	var result map[string]interface{}
	c.rstore.VisitReadonlyState(func(state state.ReadonlyState) {
		result = service.ConfigService.CompileConfig(req.Data, state, req.CommonConfigsIdList...)
	})
	return result
}

// @Summary Метод получения связей общей конфигурациями с конфигурацией модулей
// @Description Возвращает ассоциативный массив, ключами которого являются название модулей, а значения - название конфигурации модуля
// @Tags Общие конфигурации
// @Accept json
// @Produce json
// @Param body body domain.ConfigIdRequest true "id общей конфигурации"
// @Success 200 {object} domain.CommonConfigLinks
// @Router /common_config/get_links [POST]
func (c *commonConfig) GetLinks(req domain.ConfigIdRequest) domain.CommonConfigLinks {
	var result domain.CommonConfigLinks
	c.rstore.VisitReadonlyState(func(state state.ReadonlyState) {
		result = service.CommonConfig.GetCommonConfigLinks(req.Id, state)
	})
	return result
}

func NewCommonConfig(rstore *store.Store) *commonConfig {
	return &commonConfig{
		rstore: rstore,
	}
}
