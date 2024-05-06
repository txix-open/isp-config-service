package entity

import (
	"isp-config-service/entity/xtypes"
)

type EventType int

const (
	Unknown EventType = iota
	ConfigUpdated
	ModuleReady
	ModuleDisconnected
)

type Event struct {
	RowId   int                       `json:"rowid"`
	Type    EventType                 `json:"type"`
	Payload xtypes.Json[EventPayload] `json:"payload"`
}

func NewEvent(t EventType, payload EventPayload) Event {
	return Event{
		Type: t,
		Payload: xtypes.Json[EventPayload]{
			Value: payload,
		},
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
