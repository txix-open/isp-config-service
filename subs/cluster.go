package subs

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/ws"
)

func (h *socketEventHandler) applyCommandOnLeader(conn ws.Conn, cmd []byte) string {
	obj, err := h.cluster.SyncApplyOnLeader(cmd)
	if err != nil {
		var logResponse cluster.ApplyLogResponse
		logResponse.ApplyError = err.Error()
		data, err := json.Marshal(obj)
		if err != nil {
			log.Fatalf(codes.SyncApplyError, "marshaling ApplyLogResponse: %v", err)
		}
		return string(data)
	}
	data, err := json.Marshal(obj)
	if err != nil {
		log.Fatalf(codes.SyncApplyError, "marshaling ApplyLogResponse: %v", err)
	}
	return string(data)
}
