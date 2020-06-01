package service

import (
	log "github.com/integration-system/isp-log"
	"github.com/pkg/errors"
	"isp-config-service/codes"
	"isp-config-service/entity"
	"isp-config-service/holder"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

var (
	Schema schemaService
)

type schemaService struct{}

func (schemaService) HandleUpdateConfigSchemaCommand(schema entity.ConfigSchema, state state.WritableState) error {
	module := state.Modules().GetById(schema.ModuleId)
	if module == nil {
		return errors.Errorf("module with id %s not found", schema.ModuleId)
	}

	schema = state.WritableSchemas().Upsert(schema)
	if holder.ClusterClient.IsLeader() {
		// TODO handle db errors
		_, err := model.SchemaRep.Upsert(schema)
		if err != nil {
			log.WithMetadata(map[string]interface{}{
				"schema": schema,
			}).Errorf(codes.DatabaseOperationError, "upsert schema: %v", err)
		}
	}
	return nil
}
