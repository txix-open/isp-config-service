package subs

import (
	"context"
	"errors"

	etp "github.com/integration-system/isp-etp-go/v2"
	log "github.com/integration-system/isp-log"
	"isp-config-service/cluster"
	"isp-config-service/codes"
	"isp-config-service/holder"
)

func EmitConn(conn etp.Conn, event string, body []byte) {
	err := conn.Emit(context.Background(), event, body)
	if err != nil {
		log.WithMetadata(map[string]interface{}{
			"event": event,
		}).Warnf(codes.WebsocketEmitError, "emit err %v", err)
	}
}

func SyncApplyCommand(command []byte) (interface{}, error) {
	applyLogResponse, err := holder.ClusterClient.SyncApply(command)
	if err != nil {
		commandName := cluster.ParseCommand(command).String()
		log.WithMetadata(map[string]interface{}{
			"command":     string(command),
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "apply command: %v", err)
		return nil, err
	}
	if applyLogResponse != nil && applyLogResponse.ApplyError != "" {
		commandName := cluster.ParseCommand(command).String()
		log.WithMetadata(map[string]interface{}{
			"result":      string(applyLogResponse.Result),
			"applyError":  applyLogResponse.ApplyError,
			"commandName": commandName,
		}).Warnf(codes.SyncApplyError, "apply command")
		return applyLogResponse.Result, errors.New(applyLogResponse.ApplyError)
	}
	return applyLogResponse.Result, nil
}

func FormatErrorConnection(err error) []byte {
	errMap := map[string]interface{}{
		"error": err.Error(),
	}
	data, _ := json.Marshal(errMap)
	return data
}
