package domain

import "isp-config-service/entity"

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

type CreateUpdateConfigRequest struct {
	entity.Config
	Unsafe    bool
	CreateNew bool
}
