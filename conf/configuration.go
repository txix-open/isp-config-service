package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type Configuration struct {
	Database         structure.DBConfiguration      `valid:"required~Required" schema:"База данных,настройка параметров подключения к базе данных"`
	GrpcOuterAddress structure.AddressConfiguration `valid:"required~Required" json:"grpcOuterAddress"`
	ModuleName       string                         `valid:"required~Required"`
	WS               struct {
		Rest                    structure.AddressConfiguration `valid:"required~Required"`
		Grpc                    structure.AddressConfiguration `valid:"required~Required"`
		WsConnectionReadLimitKB int64                          `schema:"Максимальное количество килобайт на чтение по вебсокету,при превышении соединение закрывается с ошибкой"`
	}
	Metrics structure.MetricConfiguration
	Cluster ClusterConfiguration `valid:"required~Required"`
}

type ClusterConfiguration struct {
	InMemory              bool
	DataDir               string `valid:"required~Required"`
	BootstrapCluster      bool
	Peers                 []string `valid:"required~Required"`
	OuterAddress          string   `valid:"required~Required"`
	ConnectTimeoutSeconds int      `valid:"required~Required"`
}
