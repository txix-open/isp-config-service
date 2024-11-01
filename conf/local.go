package conf

import (
	"github.com/rqlite/rqlite/v8/auth"
	"github.com/txix-open/isp-kit/bootstrap"
)

type Local struct {
	ConfigServiceAddress     bootstrap.ConfigServiceAddr
	MaintenanceMode          bool
	Credentials              []auth.Credential
	KeepConfigVersions       int    `validate:"required"`
	InternalClientCredential string `validate:"required"`
	Rqlite                   Rqlite
}

type Rqlite struct {
	NodeId string `validate:"required"`
}
