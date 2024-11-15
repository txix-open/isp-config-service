package event

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity/xtypes"
)

type LeaderChecker interface {
	IsLeader() bool
}

type Cleaner struct {
	repo          Repo
	leaderChecker LeaderChecker
	eventTtl      time.Duration
	logger        log.Logger
}

func NewCleaner(
	repo Repo,
	leaderChecker LeaderChecker,
	eventTtl time.Duration,
	logger log.Logger,
) Cleaner {
	return Cleaner{
		repo:          repo,
		leaderChecker: leaderChecker,
		eventTtl:      eventTtl,
		logger:        logger,
	}
}

func (c Cleaner) Do(ctx context.Context) {
	ctx = log.ToContext(ctx, log.String("worker", "eventCleaner"))
	if !c.leaderChecker.IsLeader() {
		c.logger.Debug(ctx, "is not a leader, skip work")
		return
	}

	deleteBefore := time.Now().Add(-c.eventTtl)
	deleted, err := c.repo.DeleteByCreatedAt(ctx, xtypes.Time(deleteBefore))
	if err != nil {
		c.logger.Error(ctx, errors.WithMessage(err, "delete old events"))
		return
	}
	c.logger.Debug(ctx, fmt.Sprintf("delete %d old events", deleted))
}
