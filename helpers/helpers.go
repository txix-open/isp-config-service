package helpers

import (
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/txix-open/etp/v3"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity"
)

func ModuleName(conn *etp.Conn) string {
	return conn.HttpRequest().Form.Get("module_name")
}

func SplitAddress(backend entity.Backend) (cluster.AddressConfiguration, error) {
	host, port, err := net.SplitHostPort(backend.Address)
	if err != nil {
		return cluster.AddressConfiguration{}, errors.WithMessagef(err, "split backend %s address: %s", backend.ModuleId, backend.Address)
	}
	return cluster.AddressConfiguration{
		IP:   host,
		Port: port,
	}, nil
}

func LogFields(conn *etp.Conn) []log.Field {
	return []log.Field{
		log.String("moduleName", ModuleName(conn)),
		log.String("connId", strconv.FormatUint(conn.Id(), 10)),
	}
}
