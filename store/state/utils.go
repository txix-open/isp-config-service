package state

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

func StringsToMap(list []string) map[string]struct{} {
	newMap := make(map[string]struct{}, len(list))
	for _, x := range list {
		newMap[x] = struct{}{}
	}
	return newMap
}

func GenerateId() string {
	uuidV4 := uuid.NewV4()
	return uuidV4.String()
}

func GenerateDate() time.Time {
	return time.Now()
}
