package controller

import (
	"github.com/integration-system/isp-lib/utils"
	"google.golang.org/grpc/codes"
	"isp-config-service/cluster"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/store"
	"isp-config-service/store/state"
)

var CommonConfig *commonConfig

type commonConfig struct {
	rstore *store.Store
}

// GetConfigs godoc
// @Summary Метод получения объектов конфигурации по идентификаторам
// @Description Возвращает массив конфигураций по запрошенным идентификаторам (все, если массив пустой)
// @Tags Общие конфигурации
// @Accept  json
// @Produce  json
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

// CreateUpdateConfig godoc
// @Summary Метод обновления конфигурации
// @Description Если конфиг с таким id существует, то обновляет данные, если нет, то добавляет данные в базу
// В случае обновления рассылает все подключенным модулям актуальную конфигурацию
// @Tags Общие конфигурации
// @Accept  json
// @Produce  json
// @Param body body entity.Config true "объект для сохранения"
// @Success 200 {object} entity.Config
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
	err := PerformSyncApplyWithError(command, "UpsertCommonConfigCommand", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteConfigs godoc
// @Summary Метод удаления объектов конфигурации по идентификаторам
// @Description Возвращает количество удаленных модулей
// @Tags Общие конфигурации
// @Accept  json
// @Produce  json
// @Param body body []string true "массив идентификаторов конфигураций"
// @Success 200 {object} domain.DeleteResponse
// @Failure 400 {object} structure.GrpcError "если конфигурации где-то используются либо не указан объект identities"
// @Failure 500 {object} structure.GrpcError
// @Router /common_config/delete_config [POST]
func (c *commonConfig) DeleteConfigs(identities []string) (*domain.DeleteResponse, error) {
	if len(identities) == 0 {
		validationErrors := map[string]string{
			"ids": "Required",
		}
		return nil, utils.CreateValidationErrorDetails(codes.InvalidArgument, utils.ValidationError, validationErrors)
	}

	var response domain.DeleteResponse
	command := cluster.PrepareDeleteCommonConfigsCommand(identities)
	err := PerformSyncApplyWithError(command, "DeleteCommonConfigsCommand", &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func NewCommonConfig(rstore *store.Store) *commonConfig {
	return &commonConfig{
		rstore: rstore,
	}
}
