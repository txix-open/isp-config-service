package domain

type ConfigIdRequest struct {
	Id string `json:"id" valid:"required~Required"`
}

type GetByModuleIdRequest struct {
	ModuleId string `valid:"required~Required"`
}
