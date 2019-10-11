package subs

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/ws"
)

func EmitConn(conn ws.Conn, event string, args ...interface{}) {
	err := conn.Emit(event, args...)
	if err != nil {
		log.WithMetadata(map[string]interface{}{
			"event": event,
		}).Warnf(codes.SocketIoEmitError, "emit err %v", err)
	}
}

func (h *socketEventHandler) SyncApplyCommand(command []byte, commandName string) {
	applyLogResponse, err := h.cluster.SyncApply(command)
	if err != nil {
		log.Warnf(codes.SyncApplyError, "apply %s: %v", commandName, err)
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		log.WithMetadata(map[string]interface{}{
			"comment":     applyLogResponse.Comment,
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "apply command")
	}
}
