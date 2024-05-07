package controller

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/etp/v3"
	"github.com/txix-open/etp/v3/msg"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/helpers"
)

var (
	ok = []byte("ok")
)

type ModuleService interface {
	OnConnect(ctx context.Context, conn *etp.Conn, moduleName string) error
	OnDisconnect(ctx context.Context, conn *etp.Conn, moduleName string, isNormalClose bool, err error) error
	OnError(ctx context.Context, conn *etp.Conn, moduleName string, err error)
	OnModuleReady(
		ctx context.Context,
		conn *etp.Conn,
		backend cluster.BackendDeclaration,
	) error
	OnModuleRequirements(
		ctx context.Context,
		conn *etp.Conn,
		requirements cluster.ModuleRequirements,
	) error
	OnModuleConfigSchema(
		ctx context.Context,
		conn *etp.Conn,
		data cluster.ConfigData,
	) error
}

type Module struct {
	service ModuleService
	logger  log.Logger
}

func NewModule(service ModuleService, logger log.Logger) Module {
	return Module{
		service: service,
		logger:  logger,
	}
}

func (m Module) OnConnect(conn *etp.Conn) {
	ctx := conn.HttpRequest().Context()
	err := conn.HttpRequest().ParseForm()
	if err != nil {
		m.handleError(ctx, errors.WithMessage(err, "parse form"))
		_ = conn.Close()
		return
	}

	err = m.service.OnConnect(ctx, conn, helpers.ModuleName(conn))
	if err != nil {
		m.handleError(ctx, errors.WithMessage(err, "handle onConnect"))
	}
}

func (m Module) OnDisconnect(conn *etp.Conn, err error) {
	ctx := conn.HttpRequest().Context()
	handleDisconnectErr := m.service.OnDisconnect(
		ctx,
		conn,
		helpers.ModuleName(conn),
		etp.IsNormalClose(err),
		err,
	)
	if handleDisconnectErr != nil {
		m.handleError(ctx, errors.WithMessage(handleDisconnectErr, "handle onDisconnect"))
	}
}

func (m Module) OnError(conn *etp.Conn, err error) {
	m.service.OnError(conn.HttpRequest().Context(), conn, helpers.ModuleName(conn), err)
}

func (m Module) OnModuleReady(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
	backend := cluster.BackendDeclaration{}
	err := json.Unmarshal(event.Data, &backend)
	if err != nil {
		return m.handleError(ctx, errors.WithMessage(err, "unmarshal event data"))
	}

	err = m.service.OnModuleReady(ctx, conn, backend)
	if err != nil {
		return m.handleError(ctx, errors.WithMessage(err, "handle onModuleReady"))
	}

	return ok
}

func (m Module) OnModuleRequirements(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
	requirements := cluster.ModuleRequirements{}
	err := json.Unmarshal(event.Data, &requirements)
	if err != nil {
		return m.handleError(ctx, errors.WithMessage(err, "unmarshal event data"))
	}

	err = m.service.OnModuleRequirements(ctx, conn, requirements)
	if err != nil {
		return m.handleError(ctx, errors.WithMessage(err, "handle onModuleRequirements"))
	}

	return ok
}

func (m Module) OnModuleConfigSchema(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
	configData := cluster.ConfigData{}
	err := json.Unmarshal(event.Data, &configData)
	if err != nil {
		return m.handleError(ctx, errors.WithMessage(err, "unmarshal event data"))
	}

	err = m.service.OnModuleConfigSchema(ctx, conn, configData)
	if err != nil {
		return m.handleError(ctx, errors.WithMessage(err, "handle onModuleConfigSchema"))
	}

	return ok
}

func (m Module) handleError(
	ctx context.Context,
	err error,
) []byte {
	m.logger.Error(ctx, err)
	return []byte(err.Error())
}
