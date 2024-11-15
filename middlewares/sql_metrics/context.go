package sql_metrics

import (
	"context"
)

type tracerContextKey int

const (
	labelContextKey = tracerContextKey(2)
)

func OperationLabelToContext(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, labelContextKey, label)
}

func OperationLabelFromContext(ctx context.Context) string {
	value, _ := ctx.Value(labelContextKey).(string)
	return value
}
