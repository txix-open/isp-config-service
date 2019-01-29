package domain

type IdentitiesRequest struct {
	Ids []int32 `json:"ids" valid:"required~Required"`
}

type LongIdentitiesRequest struct {
	Id int64 `json:"id" valid:"required~Required"`
}

type ConfigInstanceModuleName struct {
	ModuleName   string                 `valid:"required~Required"`
	InstanceUuid string                 `valid:"required~Required"`
	ConfigData   map[string]interface{} `valid:"required~Required"`
}
