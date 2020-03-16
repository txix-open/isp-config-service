package holder

import (
	"net/http"

	etp "github.com/integration-system/isp-etp-go/v2"
	"isp-config-service/cluster"
)

var (
	ClusterClient *cluster.Client
	EtpServer     etp.Server
	HTTPServer    *http.Server
)
