package conf

import (
	"isp-config-service/service/rqlite"

	"github.com/rqlite/rqlite/v9/auth"
	"github.com/txix-open/isp-kit/bootstrap"
)

type Local struct {
	ConfigServiceAddress     bootstrap.ConfigServiceAddr
	MaintenanceMode          bool
	Credentials              []auth.Credential
	KeepConfigVersions       int    `validate:"required"`
	InternalClientCredential string `validate:"required"`
	Rqlite                   Rqlite
	Backup                   rqlite.Backup
}

type Rqlite struct {
	NodeId string `validate:"required"`
}
