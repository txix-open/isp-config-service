package service

import (
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	SchemaService schemaService
)

type schemaService struct{}

func (schemaService) HandleUpdateConfigSchema(schema entity.ConfigSchema, state state.WritableState) error {
	module := state.Modules().GetById(schema.ModuleId)
	if module == nil {
		return errors.Errorf("module with id %s not found", schema.ModuleId)
	}

	if holder.ClusterClient.IsLeader() {
		schema = state.WritableSchemas().Upsert(schema)
		_, err := model.SchemaRep.Upsert(schema)
		if err != nil {
			return err
		}
	}
	return nil
}
