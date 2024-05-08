package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/service/rqlite/db"
)

type Event struct {
	db db.DB
}

func NewEvent(db db.DB) Event {
	return Event{
		db: db,
	}
}

func (r Event) Insert(ctx context.Context, event entity.Event) error {
	query := fmt.Sprintf("insert into %s (payload) values (?)", Table("event"))
	_, err := r.db.Exec(ctx, query, event.Payload)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}
	return nil
}

func (r Event) Get(ctx context.Context, lastEventId int, limit int) ([]entity.Event, error) {
	query := fmt.Sprintf("select * from %s where id > ? order by id desc limit ?", Table("event"))
	result := make([]entity.Event, 0)
	err := r.db.Select(ctx, &result, query, lastEventId, limit)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r Event) DeleteByCreatedAt(ctx context.Context, before xtypes.Time) (int, error) {
	query := fmt.Sprintf("delete from %s where created_at < ?", Table("event"))
	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, errors.WithMessagef(err, "exec: %s", query)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WithMessage(err, "get affected rows")
	}

	return int(affected), nil
}
