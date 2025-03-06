package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"isp-config-service/entity"
	"isp-config-service/middlewares/sql_metrics"
	"isp-config-service/service/rqlite/db"
)

type Variable struct {
	db db.DB
}

func NewVariable(db db.DB) Variable {
	return Variable{
		db: db,
	}
}

func (r Variable) All(ctx context.Context) ([]entity.Variable, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.All")

	result := make([]entity.Variable, 0)
	query := fmt.Sprintf("SELECT * FROM %s order by created_at desc", Table("variable"))
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return result, nil
}

func (r Variable) GetByName(ctx context.Context, name string) (*entity.Variable, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.GetByName")

	query, args, err := squirrel.Select("*").
		From(Table("variable")).
		Where(squirrel.Eq{"name": name}).
		ToSql()
	if err != nil {
		return nil, errors.WithMessage(err, "build query")
	}

	result := entity.Variable{}
	err = r.db.SelectRow(ctx, &result, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil //nolint:nilnil
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "select: %s", query)
	}
	return &result, nil
}

func (r Variable) Insert(ctx context.Context, variable entity.Variable) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.Insert")

	query, args, err := squirrel.Insert(Table("variable")).
		Columns("name", "description", "type", "value").
		Values(variable.Name, variable.Description, variable.Type, variable.Value).
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

func (r Variable) Update(ctx context.Context, variable entity.Variable) (bool, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.Update")

	query, args, err := squirrel.Update(Table("variable")).
		SetMap(map[string]any{
			"description": variable.Description,
			"value":       variable.Value,
			"updated_at":  squirrel.Expr("unixepoch()"),
		}).Where(squirrel.Eq{
		"name": variable.Name,
	}).ToSql()
	if err != nil {
		return false, errors.WithMessage(err, "build query")
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, errors.WithMessagef(err, "exec: %s", query)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, errors.WithMessage(err, "get rows affected")
	}
	return affected > 0, nil
}

func (r Variable) Upsert(ctx context.Context, variables []entity.Variable) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.Upsert")

	requests := make([][]any, 0, len(variables))
	for _, variable := range variables {
		query, args, err := squirrel.Insert(Table("variable")).
			Columns("name", "description", "type", "value").
			Values(variable.Name, variable.Description, variable.Type, variable.Value).
			Suffix(`on conflict (name) do update 
		set description = excluded.description, type = excluded.type, value = excluded.value, updated_at = unixepoch()`,
			).ToSql()
		if err != nil {
			return errors.WithMessage(err, "build query")
		}
		requests = append(requests, db.Request(query, args...))
	}

	_, err := r.db.ExecTransaction(ctx, requests)
	if err != nil {
		return errors.WithMessagef(err, "exec transaction: %v", requests)
	}

	return nil
}

func (r Variable) UpsertLinks(ctx context.Context, configId string, links []entity.ConfigHasVariable) error {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.UpsertLinks")

	deleteQuery, args, err := squirrel.Delete(Table("config_has_variable")).Where(squirrel.Eq{
		"config_id": configId,
	}).ToSql()
	if err != nil {
		return errors.WithMessage(err, "build delete query")
	}
	requests := [][]any{
		db.Request(deleteQuery, args...),
	}
	for _, variable := range links {
		query, args, err := squirrel.Insert(Table("config_has_variable")).
			Columns("config_id", "variable_name").
			Values(variable.ConfigId, variable.VariableName).
			ToSql()
		if err != nil {
			return errors.WithMessage(err, "build query")
		}
		requests = append(requests, db.Request(query, args...))
	}

	_, err = r.db.ExecTransaction(ctx, requests)
	if err != nil {
		return errors.WithMessagef(err, "exec transaction: %v", requests)
	}

	return nil
}

func (r Variable) Delete(ctx context.Context, name string) (bool, error) {
	ctx = sql_metrics.OperationLabelToContext(ctx, "Variable.Delete")

	query, args, err := squirrel.Delete(Table("variable")).
		Where(squirrel.Eq{
			"name": name,
		}).ToSql()
	if err != nil {
		return false, errors.WithMessage(err, "build query")
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return false, errors.WithMessagef(err, "exec: %s", query)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, errors.WithMessage(err, "get rows affected")
	}
	return affected > 0, nil
}
