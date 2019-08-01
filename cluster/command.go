package cluster

import (
	"github.com/pkg/errors"
	"isp-config-service/entity"
)

type CommandType int

const (
	UpdateModule CommandType = iota
	DeleteModule
	UpdateConfiguration
	DeleteConfiguration
	UpdateSchema
	DeleteSchema
)

type Command interface {
	Apply(s State) State
}

type UpdateModuleCommand struct {
	Module entity.Module
}

type operation struct {
	Command CommandType
	Payload []byte
}

func UnmarshalCommand(data []byte) (Command, error) {
	op := operation{}
	if err := json.Unmarshal(data, &op); err != nil {
		return nil, errors.WithMessage(err, "unmarshal operation")
	}

	switch op.Command {
	case UpdateModule:

	case DeleteModule:
	case UpdateConfiguration:
	case DeleteConfiguration:
	case UpdateSchema:
	case DeleteSchema:
	}

	return nil, nil
}
