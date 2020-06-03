package domain

import (
	json2 "encoding/json"
	"isp-config-service/entity"
)

type ConfigIdRequest struct {
	Id string `json:"id" valid:"required~Required"`
}

type CompileConfigsRequest struct {
	Data                map[string]interface{}
	CommonConfigsIdList []string
}

type GetByModuleIdRequest struct {
	ModuleId string `valid:"required~Required"`
}

type GetByModuleNameRequest struct {
	ModuleName string `valid:"required~Required"`
}

type BroadcastEventRequest struct {
	ModuleNames []string `valid:"required~Required"`
	Event       string   `valid:"required~Required"`
	Payload     json2.RawMessage
}

type CreateUpdateConfigRequest struct {
	entity.Config
	Unsafe    bool
	CreateNew bool
}
