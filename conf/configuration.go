//nolint lll
package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type Configuration struct {
	Database         structure.DBConfiguration      `valid:"required~Required" schema:"База данных,настройка параметров подключения к базе данных"`
	GrpcOuterAddress structure.AddressConfiguration `valid:"required~Required" json:"grpcOuterAddress"`
	ModuleName       string                         `valid:"required~Required"`
	WS               WebService                     `valid:"required~Required" schema:"Конфигурация веб сервиса"`
	Metrics          structure.MetricConfiguration
	Cluster          ClusterConfiguration `valid:"required~Required"`
}

type ClusterConfiguration struct {
	InMemory              bool     `schema:"Хранить логи и снэпшоты только в памяти"`
	BootstrapCluster      bool     `schema:"Поднимать кластер и объявлять лидером"`
	DataDir               string   `valid:"required~Required" schema:"Пусть до директории для логов и снэпшотов рафта"`
	Peers                 []string `valid:"required~Required" schema:"Адреса всех нод в кластере,формат address:port"`
	OuterAddress          string   `valid:"required~Required" schema:"Внешний адрес ноды"`
	ConnectTimeoutSeconds int      `valid:"required~Required" schema:"Таймаут подключения"`
}

type WebService struct {
	Rest                    structure.AddressConfiguration `valid:"required~Required"`
	Grpc                    structure.AddressConfiguration `valid:"required~Required"`
	WsConnectionReadLimitKB int64                          `schema:"Максимальное количество килобайт на чтение по вебсокету,при превышении соединение закрывается с ошибкой"`
}
