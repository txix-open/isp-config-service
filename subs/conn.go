package subs

import (
	etp "github.com/integration-system/isp-etp-go/v2"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/integration-system/isp-lib/v2/utils"
	"isp-config-service/cluster"
)

type connectionState struct {
	declaration structure.BackendDeclaration
}

func IsConfigClusterNode(conn etp.Conn) bool {
	isCluster := conn.URL().Query().Get(cluster.GETParam)
	return isCluster == "true"
}

func Parameters(conn etp.Conn) (moduleName string, err error) {
	_, moduleName, err = utils.ParseParameters(conn.URL().RawQuery)
	return moduleName, err
}

func SetBackendDeclaration(conn etp.Conn, backend structure.BackendDeclaration) {
	conn.SetContext(connectionState{declaration: backend})
}

func ExtractBackendDeclaration(conn etp.Conn) (structure.BackendDeclaration, bool) {
	declaration, ok := conn.Context().(connectionState)
	return declaration.declaration, ok
}
