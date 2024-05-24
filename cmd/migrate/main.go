package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"isp-config-service/migrate/migration"
	"isp-config-service/migrations"
	"isp-config-service/repository"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: migrate postgres://username:password@localhost:5432/database_name?search_path=config_service")
		os.Exit(1)
	}
	app, err := app.New()
	if err != nil {
		panic(err)
	}
	logger := app.Logger()
	ctx := app.Context()

	pgDb, err := db.Open(ctx, os.Args[1])
	if err != nil {
		panic(err)
	}
	defer pgDb.Close()

	fileName := fmt.Sprintf("db_%d.sqlite", time.Now().Unix())
	dbClient := &db.Client{}
	dbClient.DB, err = sqlx.Open("sqlite3", fileName)
	dbClient.DB.MapperFunc(db.ToSnakeCase)
	if err != nil {
		panic(errors.WithMessage(err, "err opening db"))
	}
	defer dbClient.Close()

	mr := migration.NewRunner(migration.DialectSqlite3, migrations.Migrations, logger)
	err = mr.Run(ctx, dbClient.DB.DB)
	if err != nil {
		panic(errors.WithMessage(err, "err connecting to db"))
	}

	moduleRepo := repository.NewModule(dbClient)
	configRepo := repository.NewConfig(dbClient)
	configSchemaRepo := repository.NewConfigSchema(dbClient)
	configHistoryRepo := repository.NewConfigHistory(dbClient)

	logger.Info(ctx, "start modules migration")
	modules := make([]Module, 0)
	err = pgDb.Select(ctx, &modules, "select * from modules")
	if err != nil {
		panic(errors.WithMessage(err, "err selecting modules"))
	}
	for _, module := range modules {
		result := entity.Module{
			Id:        module.Id,
			Name:      module.Name,
			CreatedAt: xtypes.Time(module.CreatedAt),
		}
		err := moduleRepo.Upsert(ctx, result)
		if err != nil {
			panic(errors.WithMessagef(err, "err upserting module: %s", module.Name))
		}
	}
	logger.Info(ctx, "modules migrated")

	logger.Info(ctx, "start configs migration")
	configs := make([]Config, 0)
	err = pgDb.Select(ctx, &configs, "select id, module_id, version, active, created_at, updated_at, data, name from configs")
	if err != nil {
		panic(errors.WithMessage(err, "err selecting configs"))
	}
	for _, config := range configs {
		result := entity.Config{
			Id:        config.Id,
			Name:      config.Name,
			ModuleId:  config.ModuleId,
			Data:      config.Data,
			Version:   config.Version,
			Active:    xtypes.Bool(config.Active),
			AdminId:   0,
			CreatedAt: xtypes.Time(config.CreatedAt),
			UpdatedAt: xtypes.Time(config.UpdatedAt),
		}
		err := configRepo.Insert(ctx, result)
		if err != nil {
			panic(errors.WithMessagef(err, "err inserting config: %s", config.Id))
		}
	}
	logger.Info(ctx, "configs migrated")

	logger.Info(ctx, "start config schemas migration")
	schemas := make([]ConfigSchema, 0)
	err = pgDb.Select(ctx, &schemas, "select * from config_schemas")
	if err != nil {
		panic(errors.WithMessage(err, "err selecting schemas"))
	}
	for _, schema := range schemas {
		result := entity.ConfigSchema{
			Id:            schema.Id,
			ModuleId:      schema.ModuleId,
			Data:          schema.Schema,
			ModuleVersion: schema.Version,
			CreatedAt:     xtypes.Time(schema.CreatedAt),
			UpdatedAt:     xtypes.Time(schema.UpdatedAt),
		}
		err := configSchemaRepo.Upsert(ctx, result)
		if err != nil {
			panic(errors.WithMessagef(err, "err upserting config_schema: %s", schema.Id))
		}
	}
	logger.Info(ctx, "config schemas migrated")

	logger.Info(ctx, "start config history migration")
	configHistories := make([]ConfigHistory, 0)
	err = pgDb.Select(ctx, &configHistories, "select * from version_config")
	if err != nil {
		panic(errors.WithMessage(err, "err selecting configs"))
	}
	for _, history := range configHistories {
		result := entity.ConfigHistory{
			Id:        history.Id,
			ConfigId:  history.ConfigId,
			Data:      history.Data,
			Version:   history.ConfigVersion,
			AdminId:   0,
			CreatedAt: xtypes.Time(history.CreatedAt),
		}
		err := configHistoryRepo.Insert(ctx, result)
		if err != nil {
			panic(errors.WithMessagef(err, "err inserting config_history: %s", history.Id))
		}
	}
	logger.Info(ctx, "config history migrated")

	logger.Info(ctx, "done!")
}
