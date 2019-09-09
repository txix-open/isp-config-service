package subs

import (
	"isp-config-service/ws"
)

func (h *socketEventHandler) applyCommandOnLeader(conn ws.Conn, cmd []byte) string {
	_, err := h.cluster.SyncApplyOnLeader(cmd)
	if err != nil {
		return err.Error()
	}
	return Ok
}
