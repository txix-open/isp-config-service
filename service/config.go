package service

import (
	"github.com/integration-system/bellows"
	"github.com/pkg/errors"
	"isp-config-service/store/state"
)

var (
	ConfigService = configService{}
)

type configService struct{}

func (configService) GetCompiledConfig(moduleName string, state state.ReadonlyState) (map[string]interface{}, error) {
	module := state.Modules().GetByName(moduleName)
	if module == nil {
		return nil, errors.Errorf("module with name %s not found", moduleName)
	}
	config := state.Configs().GetActiveByModuleId(module.Id)
	if config == nil {
		return nil, errors.Errorf("no active configs for moduleId %s", module.Id)
	}
	commonConfigs := state.CommonConfigs().GetByIds(config.CommonConfigs)
	configsToMerge := make([]map[string]interface{}, 0, len(commonConfigs))
	for _, common := range commonConfigs {
		configsToMerge = append(configsToMerge, common.Data)
	}
	configsToMerge = append(configsToMerge, config.Data)

	resultData := mergeNestedMaps(configsToMerge...)
	return resultData, nil
}

func mergeNestedMaps(maps ...map[string]interface{}) map[string]interface{} {
	if len(maps) == 1 {
		return maps[0]
	}
	result := bellows.Flatten(maps[0])
	for i := 1; i < len(maps); i++ {
		newFlatten := bellows.Flatten(maps[i])
		for k, v := range newFlatten {
			result[k] = v
		}
	}
	return bellows.Expand(result).(map[string]interface{})
}
