package entity

import "time"

type Module struct {
	tableName          string    `sql:"?db_schema.modules" json:"-"`
	Id                 string    `json:"id" valid:"required~Required"`
	Name               string    `json:"name" valid:"required~Required"`
	Active             bool      `json:"active" sql:"-"`
	CreatedAt          time.Time `json:"createdAt" sql:",null"`
	LastConnectedAt    time.Time `json:"lastConnectedAt"`
	LastDisconnectedAt time.Time `json:"lastDisconnectedAt"`
}
