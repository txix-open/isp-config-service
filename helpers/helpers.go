package helpers

import (
	"github.com/txix-open/etp/v3"
)

func ModuleName(conn *etp.Conn) string {
	_ = conn.HttpRequest().ParseForm()
	return conn.HttpRequest().Form.Get("module_name")
}
