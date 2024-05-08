package event

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/service/rqlite/db"
)

const (
	limit = 100
)

type HandlerRepo interface {
	Handle(ctx context.Context, events []entity.Event)
}

type Repo interface {
	Get(ctx context.Context, lastRowId int, limit int) ([]entity.Event, error)
	DeleteByCreatedAt(ctx context.Context, before xtypes.Time) (int, error)
}

type Worker struct {
	repo        Repo
	compactor   Compactor
	handler     HandlerRepo
	lastEventId int
	logger      log.Logger
}

func NewWorker(repo Repo, handler HandlerRepo, logger log.Logger) *Worker {
	return &Worker{
		repo:      repo,
		handler:   handler,
		compactor: NewCompactor(),
		logger:    logger,
	}
}

func (w *Worker) Do(ctx context.Context) {
	ctx = log.ToContext(ctx, log.String("worker", "eventReader"))
	err := w.do(ctx)
	if err != nil {
		w.logger.Error(ctx, err)
	}
}

func (w *Worker) do(ctx context.Context) error {
	getEventCtx := db.NoneConsistency().ToContext(ctx)
	events, err := w.repo.Get(getEventCtx, w.lastEventId, limit)
	if err != nil {
		return errors.WithMessage(err, "get new events")
	}
	if len(events) == 0 {
		return nil
	}

	events = w.compactor.Compact(events)
	w.handler.Handle(ctx, events)
	w.lastEventId = events[len(events)-1].Id

	return nil
}
