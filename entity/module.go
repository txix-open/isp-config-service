package entity

import (
	"isp-config-service/entity/xtypes"
)

// nolint:tagliatelle
type Module struct {
	Id                 string       `json:"id"`
	Name               string       `json:"name"`
	LastConnectedAt    *xtypes.Time `json:"last_connected_at"`
	LastDisconnectedAt *xtypes.Time `json:"last_disconnected_at"`
	CreatedAt          xtypes.Time  `json:"created_at"`
}
