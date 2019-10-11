package state

import (
	"github.com/integration-system/bellows"
	uuid "github.com/satori/go.uuid"
)

func StringsToMap(list []string) map[string]struct{} {
	newMap := make(map[string]struct{}, len(list))
	for _, x := range list {
		newMap[x] = struct{}{}
	}
	return newMap
}

func MergeNestedMaps(maps ...map[string]interface{}) map[string]interface{} {
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

func GenerateId() string {
	uuidV4 := uuid.NewV4()
	return uuidV4.String()
}
