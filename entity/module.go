package entity

import "time"

type Module struct {
	tableName          string    `sql:"?db_schema.modules" json:"-"`
	Id                 int32     `json:"id"`
	Uuid               string    `json:"uuid" valid:"required~Required,uuid~must be a valid uuid"`
	Name               string    `json:"name" valid:"required~Required"`
	Active             bool      `json:"active" sql:"-"`
	CreatedAt          time.Time `json:"createdAt" sql:",null"`
	LastConnectedAt    time.Time `json:"lastConnectedAt"`
	LastDisconnectedAt time.Time `json:"lastDisconnectedAt"`
}
