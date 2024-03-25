package subs

import (
	"fmt"

	etp "github.com/integration-system/isp-etp-go/v2"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/holder"
)

func (h *SocketEventHandler) applyCommandOnLeader(_ etp.Conn, cmd []byte) (data []byte) {
	cmdCopy := make([]byte, len(cmd))
	copy(cmdCopy, cmd)
	obj, err := holder.ClusterClient.SyncApplyOnLeader(cmdCopy)
	if err != nil {
		var logResponse cluster.ApplyLogResponse
		logResponse.ApplyError = fmt.Sprintf("SyncApplyOnLeader: %v", err)
		data, err = json.Marshal(logResponse)
		if err != nil {
			log.Fatalf(codes.SyncApplyError, "marshaling ApplyLogResponse: %v", err)
		}
		return data
	}
	data, err = json.Marshal(obj)
	if err != nil {
		log.Fatalf(codes.SyncApplyError, "marshaling ApplyLogResponse: %v", err)
	}
	return data
}
