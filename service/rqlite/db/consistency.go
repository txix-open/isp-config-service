package db

import (
	"context"
	"time"
)

type consistencyCtxKey struct{}

var (
	consistencyCtxValue = consistencyCtxKey{}
)

func ConsistencyFromContext(ctx context.Context) Consistency {
	c, ok := ctx.Value(consistencyCtxValue).(Consistency)
	if !ok {
		return Consistency{}
	}
	return c
}

type ConsistencyLevel string

const (
	ConsistencyLevelNone   ConsistencyLevel = "none"
	ConsistencyLevelWeak   ConsistencyLevel = "weak"
	ConsistencyLevelStrong ConsistencyLevel = "strong"
)

type Consistency struct {
	level           ConsistencyLevel
	freshness       time.Duration
	freshnessStrict bool
}

func (c Consistency) WithLevel(level ConsistencyLevel) Consistency {
	c.level = level
	return c
}

func (c Consistency) WithFreshness(freshness time.Duration) Consistency {
	c.freshness = freshness
	return c
}

func (c Consistency) WithFreshnessStrict(freshnessStrict bool) Consistency {
	c.freshnessStrict = freshnessStrict
	return c
}
func (c Consistency) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, consistencyCtxValue, c)
}

func (c Consistency) appendParams(params map[string]any) {
	if c.level != "" {
		params["level"] = c.level
	}
	if c.freshness > 0 {
		params["freshness"] = c.freshness.String()
	}
	if c.freshnessStrict {
		params["freshness_strict"] = c.freshnessStrict
	}
}

func NoneConsistency() Consistency {
	return Consistency{
		level: ConsistencyLevelNone,
	}
}

func WeakConsistency() Consistency {
	return Consistency{
		level: ConsistencyLevelWeak,
	}
}

func StrongConsistency() Consistency {
	return Consistency{
		level: ConsistencyLevelStrong,
	}
}
