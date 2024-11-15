package backend

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
)

type Service interface {
	ClearPhantomBackends(ctx context.Context) (int, error)
}

type Cleaner struct {
	service Service
	logger  log.Logger
}

func NewCleaner(
	service Service,
	logger log.Logger,
) Cleaner {
	return Cleaner{
		service: service,
		logger:  logger,
	}
}

func (c Cleaner) Do(ctx context.Context) {
	ctx = log.ToContext(ctx, log.String("worker", "phantomBackendCleaner"))

	deleted, err := c.service.ClearPhantomBackends(ctx)
	if err != nil {
		c.logger.Error(ctx, errors.WithMessage(err, "clear phantom backends"))
		return
	}

	c.logger.Debug(ctx, fmt.Sprintf("delete %d phantom backends", deleted))
}
