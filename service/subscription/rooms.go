package subscription

import (
	"fmt"
)

func ConfigChangingRoom(moduleId string) string {
	return fmt.Sprintf("config_changing_room.%s", moduleId)
}

func BackendsChangingRoom(moduleId string) string {
	return fmt.Sprintf("backends_changing_room.%s", moduleId)
}

func RoutingChangingRoom() string {
	return "routing_changing_room"
}
