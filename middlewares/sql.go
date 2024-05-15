package middlewares

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics"
	"isp-config-service/middlewares/sql_metrics"
)

func SqlOperationMiddleware() httpcli.Middleware {
	//nolint:promlinter
	sqlQueryDuration := metrics.GetOrRegister(metrics.DefaultRegistry, prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "sql",
		Name:       "query_duration_ms",
		Help:       "The latencies of sql query",
		Objectives: metrics.DefaultObjectives,
	}, []string{"operation"}))
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			label := sql_metrics.OperationLabelFromContext(ctx)
			now := time.Now()
			resp, err := next.RoundTrip(ctx, request)
			if label != "" {
				duration := time.Since(now)
				sqlQueryDuration.WithLabelValues(label).Observe(metrics.Milliseconds(duration))
			}
			return resp, err //nolint:wrapcheck
		})
	}
}
