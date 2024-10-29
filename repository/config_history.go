package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/middlewares/sql_metrics"
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
	ctx = sql_metrics.OperationLabelToContext(ctx, "ConfigHistory.GetByConfigId")

	query := fmt.Sprintf(`select * from %s where config_id = ? order by version desc`, Table("config_history"))
	result := make([]entity.ConfigHistory, 0)
	err := r.db.Select(ctx, &result, query, configId)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r ConfigHistory) Delete(ctx context.Context, id string) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "ConfigHistory.Delete")

	query := fmt.Sprintf(`delete from %s where id = ?`, Table("config_history"))
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}
	return nil
}

func (r ConfigHistory) Insert(ctx context.Context, history entity.ConfigHistory) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "ConfigHistory.Insert")

	query, args, err := squirrel.Insert(Table("config_history")).
		Columns("id", "config_id", "data", "version", "admin_id").
		Values(history.Id, history.ConfigId, history.Data, history.Version, history.AdminId).
		ToSql()
	if err != nil {
		return errors.WithMessage(err, "build query")
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.WithMessagef(err, "exec: %s", query)
	}

	return nil
}

func (r ConfigHistory) DeleteOld(ctx context.Context, configId string, keepVersions int) (int, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "ConfigHistory.DeleteOld")

	query := fmt.Sprintf(`delete from %s where config_id = ? and id not in (
                   select id from %s where config_id = ? order by version desc limit ?
            	)`, Table("config_history"), Table("config_history"))
	result, err := r.db.Exec(ctx, query, configId, configId, keepVersions)
	if err != nil {
		return 0, errors.WithMessagef(err, "exec: %s", query)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WithMessage(err, "get rows affected")
	}

	return int(rowsAffected), nil
}
