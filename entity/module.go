package entity

import (
	"time"
)

type Module struct {
	Id                 int32     `json:"id"`
	InstanceId         int32     `json:"instanceId" valid:"required~Required"`
	Name               string    `json:"name" valid:"required~Required"`
	Active             bool      `json:"active" sql:",null"`
	CreatedAt          time.Time `json:"createdAt" sql:",null"`
	LastConnectedAt    time.Time `json:"lastConnectedAt"`
	LastDisconnectedAt time.Time `json:"lastDisconnectedAt"`
}

type ModuleInstanceIdentity struct {
	Name string `valid:"required~Required"`
	Uuid string `valid:"required~Required,uuid~must be a valid uuid"`
}
