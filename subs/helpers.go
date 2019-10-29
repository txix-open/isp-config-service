package subs

import (
	"errors"
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

func SyncApplyCommand(command []byte, commandName string) (interface{}, error) {
	applyLogResponse, err := holder.ClusterClient.SyncApply(command)
	if err != nil {
		return nil, err
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		log.WithMetadata(map[string]interface{}{
			"result":      string(applyLogResponse.Result),
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "apply command")
		return applyLogResponse.Result, errors.New(applyLogResponse.ApplyError)
	}
	return applyLogResponse.Result, nil
}
