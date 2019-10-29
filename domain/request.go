package domain

type ConfigIdRequest struct {
	Id string `json:"id" valid:"required~Required"`
}

type GetByModuleIdRequest struct {
	ModuleId string `valid:"required~Required"`
}

type GetByModuleNameRequest struct {
	ModuleName string `valid:"required~Required"`
}
