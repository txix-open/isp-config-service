-- +goose Up
set migration.internal_db_host = 'isp-pgsql';
set migration.internal_db_port = '5432';
set migration.internal_db_dbname = 'isp-test';
set migration.internal_db_user = 'isp-test';
set migration.internal_db_pass = '123321';


WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-async-api', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ((format('{
              "limits": {
                "totalWorkers": 4,
                "jobExecutionAttempts": 1
              },
              "metrics": {
                "address": {
                  "ip": "0.0.0.0",
                  "path": "/metrics",
                  "port": "9576"
                },
                "gc": true,
                "memory": true
              },
              "database": {
                "password": "%s",
                "username": "%s",
                "createSchema": true,
                "port": "%s",
                "schema": "mdm_async_api_service",
                "address": "%s",
                "database": "%s"
              },
              "execution": {
                "acquiredJobTTLInSec": 300,
                "storeEntriesTTLInSec": 1800,
                "purgeStoreTimeoutInSec": 600,
                "jobStealingRequestTimeoutInSec": 5,
                "scrollTTLInSec": 60
              }
            }',
    current_setting('migration.internal_db_pass'),
    current_setting('migration.internal_db_user'),
    current_setting('migration.internal_db_port'),
    current_setting('migration.internal_db_host'),
    current_setting('migration.internal_db_dbname')
)) :: jsonb));

-- +goose Down

