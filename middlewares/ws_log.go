package middlewares

import (
	"context"
	"github.com/txix-open/etp/v4"

	"github.com/txix-open/etp/v4/msg"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/helpers"
)

type EtpMiddleware func(next etp.Handler) etp.Handler

func EtpLogger(logger log.Logger) EtpMiddleware {
	return func(next etp.Handler) etp.Handler {
		return etp.HandlerFunc(func(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
			fields := append(
				helpers.LogFields(conn),
				log.String("event", event.Name),
			)
			logger.Debug(ctx, "event received", fields...)
			return next.Handle(ctx, conn, event)
		})
	}
}

//nolint:ireturn
func EtpChain(root etp.Handler, middlewares ...EtpMiddleware) etp.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		root = middlewares[i](root)
	}
	return root
}
