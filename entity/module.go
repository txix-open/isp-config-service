package entity

import (
	"time"
)

type Module struct {
	Id                 string     `json:"id"`
	Name               string     `json:"name"`
	LastConnectedAt    *time.Time `json:"last_connected_at"`
	LastDisconnectedAt *time.Time `json:"last_disconnected_at"`
	CreatedAt          time.Time  `json:"created_at"`
}
