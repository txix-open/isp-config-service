package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
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
	query := fmt.Sprintf("insert into %s (type, payload) values (?, ?)", Table("event"))
	_, err := r.db.Exec(ctx, query, event.Type, event.Payload)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}
	return nil
}

func (r Event) Get(ctx context.Context, lastRowId int, limit int) ([]entity.Event, error) {
	query := fmt.Sprintf("select * from %s where rowid > ? order by rowid desc limit ?", Table("event"))
	result := make([]entity.Event, 0)
	err := r.db.Select(ctx, &result, query, lastRowId, limit)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r Event) DeleteByCreatedAt(ctx context.Context, before int) (int, error) {
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
