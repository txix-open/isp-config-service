package entity

import (
	"time"
)

type Instance struct {
	Id   int32  `json:"id"`
	Uuid string `json:"uuid" valid:"required~Required,uuid~must be a valid uuid"`
	Name string `json:"name" valid:"required~Required"`
	// CreatedAtJson utils.JSONTime `json:"createdAt" sql:"-"`
	CreatedAt time.Time `json:"createdAt" sql:",null"`
}
