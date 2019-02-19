-- +goose Up
set migration.ds_db_host = 'isp-pgsql';
set migration.ds_db_port = '5432';
set migration.ds_db_dbname = 'isp-test';
set migration.ds_db_user = 'isp-test';
set migration.ds_db_pass = '123321';

set migration.internal_db_host = 'isp-pgsql';
set migration.internal_db_port = '5432';
set migration.internal_db_dbname = 'isp-test';
set migration.internal_db_user = 'isp-test';
set migration.internal_db_pass = '123321';


WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-dumper', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ((format('{
  "limits": {
    "rowsInFile": 1000000,
    "fetchBatchSize": 1000,
    "convertBatchSize": 300
  },
  "metrics": {
    "gc": true,
    "memory": true,
    "address": {
      "port": "9575",
      "ip": "0.0.0.0",
      "path": "/metrics"
    }
  },
  "dataSourceDb": {
    "password": "%s",
    "username": "%s",
    "createSchema": false,
    "port": "%s",
    "schema": "mdm_service",
    "address": "%s",
    "database": "%s"
  },
  "internalDb": {
      "password": "%s",
      "username": "%s",
      "createSchema": true,
      "port": "%s",
      "schema": "mdm_dumper_service",
      "address": "%s",
      "database": "%s"
    },
  "schedules": [],
  "workerCount": 1,
  "dumpDirectoryPath": "/opt/msp/msp-mdm-dumper-service/dump"
}', current_setting('migration.ds_db_pass'),
    current_setting('migration.ds_db_user'),
    current_setting('migration.ds_db_port'),
    current_setting('migration.ds_db_host'),
    current_setting('migration.ds_db_dbname'),

    current_setting('migration.internal_db_pass'),
    current_setting('migration.internal_db_user'),
    current_setting('migration.internal_db_port'),
    current_setting('migration.internal_db_host'),
    current_setting('migration.internal_db_dbname')
)) :: jsonb));

-- +goose Down

