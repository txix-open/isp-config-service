package module

import (
	"context"
	"strings"
	"time"

	"github.com/txix-open/etp/v4"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/helpers"
)

const (
	emitEventTimeout = 5 * time.Second
)

type Emitter struct {
	logger log.Logger
}

func NewEmitter(logger log.Logger) Emitter {
	return Emitter{
		logger: logger,
	}
}

func (s Emitter) Emit(
	ctx context.Context,
	conn *etp.Conn,
	event string,
	data []byte,
) {
	ctx, cancel := context.WithTimeout(ctx, emitEventTimeout)
	defer cancel()

	fields := append(helpers.LogFields(conn), log.String("event", event))
	if strings.HasSuffix(event, cluster.ModuleConnectionSuffix) {
		fields = append(fields, log.ByteString("data", data))
	}

	s.logger.Debug(
		ctx,
		"emit event",
		fields...,
	)

	err := conn.Emit(ctx, event, data)

	if err != nil {
		err := errors.WithMessagef(
			err,
			"emit event '%s', to %s module, connId: %s",
			event, helpers.ModuleName(conn), conn.Id(),
		)
		s.logger.Error(ctx, err)
	}
}
