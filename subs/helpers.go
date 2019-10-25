package subs

import (
	log "github.com/integration-system/isp-log"
	"isp-config-service/codes"
	"isp-config-service/holder"
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

func SyncApplyCommand(command []byte, commandName string) {
	applyLogResponse, err := holder.ClusterClient.SyncApply(command)
	if err != nil {
		log.Warnf(codes.SyncApplyError, "apply %s: %v", commandName, err)
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		log.WithMetadata(map[string]interface{}{
			"result":      string(applyLogResponse.Result),
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "apply command")
	}
}
