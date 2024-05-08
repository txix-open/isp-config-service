package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/service/rqlite/db"
)

type ConfigHistory struct {
	db db.DB
}

func NewConfigHistory(db db.DB) ConfigHistory {
	return ConfigHistory{
		db: db,
	}
}

func (r ConfigHistory) GetByConfigId(ctx context.Context, configId string) ([]entity.ConfigHistory, error) {
	query := fmt.Sprintf(`select * from %s where config_id = ? order by version desc`, Table("config_history"))
	result := make([]entity.ConfigHistory, 0)
	err := r.db.Select(ctx, &result, query, configId)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r ConfigHistory) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`delete from %s where id = ?`, Table("config_history"))
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}
	return nil
}
