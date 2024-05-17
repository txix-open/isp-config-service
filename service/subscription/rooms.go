package subscription

import (
	"fmt"
)

func ConfigChangingRoom(moduleId string) string {
	return fmt.Sprintf("config_changing_room.%s", moduleId)
}

func BackendsChangingRoom(moduleName string) string {
	return fmt.Sprintf("backends_changing_room.%s", moduleName)
}

func RoutingChangingRoom() string {
	return "routing_changing_room"
}
