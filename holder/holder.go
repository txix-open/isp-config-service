package holder

import (
	"isp-config-service/cluster"
	"isp-config-service/ws"
	"net/http"
)

var (
	ClusterClient *cluster.Client
	Socket        *ws.WebsocketServer
	HttpServer    *http.Server
)
