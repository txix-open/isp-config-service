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

// Module leader recompute reasons. They tell the log which change of the cluster put the event in the queue,
// and only that: the leader itself is taken from the module instances every time it is needed.
const (
	ModuleLeaderReasonReady      = "ready"
	ModuleLeaderReasonDisconnect = "disconnect"
)

func (e Event) Key() string {
	switch {
	case e.Payload.Value.ConfigUpdated != nil:
		return fmt.Sprintf("config_updated_%s", e.Payload.Value.ConfigUpdated.ModuleId)
	case e.Payload.Value.ModuleReady != nil:
		return fmt.Sprintf("module_lifecycle_%s", e.Payload.Value.ModuleReady.ModuleId)
	case e.Payload.Value.ModuleDisconnected != nil:
		return fmt.Sprintf("module_lifecycle_%s", e.Payload.Value.ModuleDisconnected.ModuleId)
	case e.Payload.Value.ModuleLeaderRecompute != nil:
		// The leader recompute keeps its own key: the compactor collapses ready and disconnected into one
		// lifecycle event per module, and the recompute must not be swallowed by that collapse.
		return fmt.Sprintf("module_leader_%s", e.Payload.Value.ModuleLeaderRecompute.ModuleId)
	default:
		return ""
	}
}

// nolint:tagliatelle
type EventPayload struct {
	ConfigUpdated      *PayloadConfigUpdated      `json:",omitempty"`
	ModuleReady        *PayloadModuleReady        `json:",omitempty"`
	ModuleDisconnected *PayloadModuleDisconnected `json:",omitempty"`
	// ModuleLeaderRecompute is internal: no module receives it as a WebSocket event.
	ModuleLeaderRecompute *PayloadModuleLeaderRecompute `json:",omitempty"`
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

// PayloadModuleLeaderRecompute asks the config service to look at the module leader again.
//
// PreviousLeader is the address the leader had before the change, and Leader is the address derived from the
// module instances after it, empty when the module has no leader left. Both are assembled where the change
// happens, together with the events of that change, so the event worker reads nothing from the database: an
// extra read there races with the events of the same cluster and slows the whole queue down. The leader
// address the modules see is still elected by the read path when it serves a request.
//
// nolint:tagliatelle
type PayloadModuleLeaderRecompute struct {
	ModuleId   string
	ModuleName string
	Reason     string
	// PreviousLeader is empty when the module had no leader.
	PreviousLeader string
	// Leader is empty when the module has no leader after the change.
	Leader string
}
