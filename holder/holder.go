package holder

import (
	"isp-config-service/cluster"
	"isp-config-service/ws"
	"net/http"
)

var (
	ClusterClient *cluster.ClusterClient
	Socket        *ws.WebsocketServer
	HttpServer    *http.Server
)
