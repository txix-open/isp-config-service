package conf

import (
	"github.com/rqlite/rqlite/v8/auth"
	"github.com/txix-open/isp-kit/bootstrap"
)

type Local struct {
	ConfigServiceAddress     bootstrap.ConfigServiceAddr
	MaintenanceMode          bool
	Credentials              []auth.Credential
	InternalClientCredential string `validate:"required"`
}
