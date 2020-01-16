package store

import (
	"github.com/pkg/errors"
	"isp-config-service/model"
	"isp-config-service/store/state"
)

func NewStateFromRepository() (*state.State, error) {
	configs, err := model.ConfigRep.Snapshot()
	if err != nil {
		return nil, errors.WithMessage(err, "restore configs")
	}
	schemas, err := model.SchemaRep.Snapshot()
	if err != nil {
		return nil, errors.WithMessage(err, "load config schemas")
	}
	modules, err := model.ModuleRep.Snapshot()
	if err != nil {
		return nil, errors.WithMessage(err, "load modules registry")
	}
	commonConfigs, err := model.CommonConfigRep.Snapshot()
	if err != nil {
		return nil, errors.WithMessage(err, "load common configs")
	}

	st := state.NewStateFromSnapshot(configs, schemas, modules, commonConfigs)
	return st, nil
}
