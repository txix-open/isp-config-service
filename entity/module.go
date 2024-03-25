package entity

import "time"

type Module struct {
	//nolint
	tableName          string    `pg:"?db_schema.modules" json:"-"`
	Id                 string    `json:"id"`
	Name               string    `json:"name" valid:"required~Required"`
	CreatedAt          time.Time `json:"createdAt" pg:",null"`
	LastConnectedAt    time.Time `json:"lastConnectedAt"`
	LastDisconnectedAt time.Time `json:"lastDisconnectedAt"`
}
