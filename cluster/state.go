package cluster

import "isp-config-service/entity"

type State struct {
	Configurations map[int64]entity.Config
	Modules        map[int32]entity.Module
	Instances      map[int32]entity.Instance
	Schemas        map[int32]entity.ConfigSchema
}
