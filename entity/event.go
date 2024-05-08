package entity

import (
	"fmt"

	"isp-config-service/entity/xtypes"
)

type Event struct {
	Id      int                       `json:"id"`
	Payload xtypes.Json[EventPayload] `json:"payload"`
}

func NewEvent(payload EventPayload) Event {
	return Event{
		Payload: xtypes.Json[EventPayload]{
			Value: payload,
		},
	}
}

func (e Event) Key() string {
	switch {
	case e.Payload.Value.ConfigUpdated != nil:
		return fmt.Sprintf("config_updated_%s", e.Payload.Value.ConfigUpdated.ModuleId)
	case e.Payload.Value.ModuleReady != nil:
		return fmt.Sprintf("module_lifecycle_%s", e.Payload.Value.ModuleReady.ModuleId)
	case e.Payload.Value.ModuleDisconnected != nil:
		return fmt.Sprintf("module_lifecycle_%s", e.Payload.Value.ModuleDisconnected.ModuleId)
	default:
		return ""
	}
}

type EventPayload struct {
	ConfigUpdated      *PayloadConfigUpdated      `json:",omitempty"`
	ModuleReady        *PayloadModuleReady        `json:",omitempty"`
	ModuleDisconnected *PayloadModuleDisconnected `json:",omitempty"`
}

type PayloadConfigUpdated struct {
	ModuleId string
}

type PayloadModuleReady struct {
	ModuleId string
}

type PayloadModuleDisconnected struct {
	ModuleId string
}
