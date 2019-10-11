package service

import (
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var SchemaService schemaService

type schemaService struct{}

func (schemaService) HandleUpdateConfigSchema(schema entity.ConfigSchema, state state.State) (state.State, error) {
	_, ok := state.GetModuleById(schema.ModuleId)
	if !ok {
		return state, errors.Errorf("module with id %s not found", schema.ModuleId)
	}

	//todo add check is leader?
	schema = state.UpdateSchema(schema)
	_, err := model.SchemaRep.Upsert(schema)
	if err != nil {
		return state, err
	}
	return state, nil
}
