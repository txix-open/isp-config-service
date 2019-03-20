package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type Configuration struct {
	Database         structure.DBConfiguration      `valid:"required~Required"`
	GrpcOuterAddress structure.AddressConfiguration `valid:"required~Required" json:"grpcOuterAddress"`
	ModuleName       string                         `valid:"required~Required"`
	WS               struct {
		Rest structure.AddressConfiguration `valid:"required~Required"`
		Grpc structure.AddressConfiguration `valid:"required~Required"`
	}
	Metrics structure.MetricConfiguration
}
