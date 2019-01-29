-- +goose Up
set migration.db_host = 'msk-pgsql.isp.mobi';
set migration.db_port = '5432';
set migration.db_dbname = 'isp-alpha';
set migration.db_user = 'isp-alpha';
set migration.db_pass = 'JxY96xLF8zfh3g8EWvSDtLUBw';

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'template', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{"database": {' ||
   '"port": "' || current_setting('migration.db_port') || '", ' ||
    '"schema": "template_service", ' ||
     '"address": "' || current_setting('migration.db_host') || '", ' ||
      '"database": "' || current_setting('migration.db_dbname') || '", ' ||
       '"password": "' || current_setting('migration.db_pass') || '", ' ||
        '"username": "' || current_setting('migration.db_user') || '", ' ||
         '"createSchema": true}, "defaultLanguage": "en", ' ||
         '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9567"}}' ||
          '}')::jsonb);

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'message-bundle', 't') RETURNING id
)
INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{"database": {' ||
   '"port": "' || current_setting('migration.db_port') || '", ' ||
    '"schema": "message_bundle_service", ' ||
     '"address": "' || current_setting('migration.db_host') || '", ' ||
      '"database": "' || current_setting('migration.db_dbname') || '", ' ||
       '"password": "' || current_setting('migration.db_pass') || '", ' ||
        '"username": "' || current_setting('migration.db_user') || '", ' ||
         '"createSchema": true}, "defaultLanguage": "en", ' ||
          '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9560"}}' ||
           '}')::jsonb);

-- +goose Down
