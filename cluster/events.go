package cluster

import json2 "encoding/json"

type ModuleConnected struct {
	ModuleName string
}

type ApplyLogResponse struct {
	ApplyError string
	Result     json2.RawMessage
}

func (a ApplyLogResponse) Error() string {
	return a.ApplyError
}
