package controller

import (
	"context"

	"github.com/txix-open/etp/v3"
	"github.com/txix-open/etp/v3/msg"
)

type Module struct {
}

func NewModule() Module {
	return Module{}
}

func (m Module) OnConnect(conn *etp.Conn) {

}

func (m Module) OnDisconnect(ctx *etp.Conn, conn error) {

}

func (m Module) OnError(conn *etp.Conn, err error) {

}

func (m Module) OnModuleReady(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
	return nil
}

func (m Module) OnModuleRequirements(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
	return nil
}

func (m Module) OnModuleConfigSchema(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
	return nil
}
