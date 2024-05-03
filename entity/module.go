package entity

import (
	"time"
)

type Module struct {
	Id                 string
	Name               string
	LastConnectedAt    time.Time
	LastDisconnectedAt time.Time
	Created            time.Time
}
