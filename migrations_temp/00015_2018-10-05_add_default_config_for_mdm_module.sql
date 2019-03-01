-- +goose Up
set migrationWrite.db_host = 'isp-pgsql';
set migrationWrite.db_port = '5432';
set migrationWrite.db_dbname = 'isp-test';
set migrationWrite.db_user = 'isp-test';
set migrationWrite.db_pass = '123321';

set migrationRead.db_host = 'isp-pgsql';
set migrationRead.db_port = '5432';
set migrationRead.db_dbname = 'isp-test';
set migrationRead.db_user = 'isp-test';
set migrationRead.db_pass = '123321';

set "rabbit.host" = 'isp-rabbit';
set "rabbit.port" = '5672';
set "rabbit.user" = 'guest';
set "rabbit.password" = 'guest';

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ((format('{
  "queue": {
    "exchange": "notifications",
    "notifierQueue": "notifier.queue",
    "searcherQueue": "searcher.queue"
  },
  "rabbit": {
    "user": "%s",
    "address": {
      "ip": "%s",
      "port": "%s"
    },
    "password": "%s"
  },
  "metrics": {
    "gc": true,
    "memory": true,
    "address": {
      "ip": "0.0.0.0",
      "path": "/metrics",
      "port": "9568"
    }
  },
  "databaseRead": {
    "password": "%s",
    "username": "%s",
    "createSchema": true,
    "port": "%s",
    "schema": "mdm_service",
    "address": "%s",
    "database": "%s"
  },
  "databaseWrite": {
    "password": "%s",
    "username": "%s",
    "createSchema": true,
    "port": "%s",
    "schema": "mdm_service",
    "address": "%s",
    "database": "%s"
  },
  "externalIdListLimit": 10000
}', current_setting('rabbit.user'),
    current_setting('rabbit.host'),
    current_setting('rabbit.port'),
    current_setting('rabbit.password'),

    current_setting('migrationRead.db_pass'),
    current_setting('migrationRead.db_user'),
    current_setting('migrationRead.db_port'),
    current_setting('migrationRead.db_host'),
    current_setting('migrationRead.db_dbname'),

    current_setting('migrationWrite.db_pass'),
    current_setting('migrationWrite.db_user'),
    current_setting('migrationWrite.db_port'),
    current_setting('migrationWrite.db_host'),
    current_setting('migrationWrite.db_dbname')

)) :: jsonb));

-- +goose Down

