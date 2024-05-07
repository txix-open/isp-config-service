package event

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity/xtypes"
)

type Cleaner struct {
	repo     Repo
	eventTtl time.Duration
	logger   log.Logger
}

func NewCleaner(repo Repo, eventTtl time.Duration, logger log.Logger) Cleaner {
	return Cleaner{
		repo:     repo,
		eventTtl: eventTtl,
		logger:   logger,
	}
}

func (c Cleaner) Do(ctx context.Context) {
	deleteBefore := time.Now().Add(-c.eventTtl)
	deleted, err := c.repo.DeleteByCreatedAt(ctx, xtypes.Time{Value: deleteBefore})
	if err != nil {
		c.logger.Error(ctx, errors.WithMessage(err, "delete old events"))
		return
	}
	c.logger.Debug(ctx, fmt.Sprintf("delete %d old events", deleted))
}
