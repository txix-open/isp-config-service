-- +goose Up
set migration.db_host = 'isp-pgsql';
set migration.db_port = '5432';
set migration.db_dbname = 'isp-test';
set migration.db_user = 'isp-test';
set migration.db_pass = '123321';

set migration.redis_host = 'isp-redis';
set migration.redis_port = '6379';

-- auth
INSERT INTO configs VALUES (1, 1, NULL, 't', NULL, NULL, ('{"redis": {' ||
 '"ip": "' || current_setting('migration.redis_host') || '", ' ||
 '"port": ' || current_setting('migration.redis_port') ||
 '}, "header": {"token": {"user": "X-USER-TOKEN", "device": "X-DEVICE-TOKEN", "application": "X-APPLICATION-TOKEN"}, "identity": {"user": "X-USER-IDENTITY", "device": "X-DEVICE-IDENTITY", "domain": "X-DOMAIN-IDENTITY", "system": "X-SYSTEM-IDENTITY", "service": "X-SERVICE-IDENTITY", "application": "X-APPLICATION-IDENTITY"}}, "unauthorized.html": "<HTML><BODY><H1>401 Unauthorized BAD CREDENTIALS</H1></BODY></HTML>"}')::jsonb);
-- converter
INSERT INTO configs VALUES (2, 2, NULL, 't', NULL, NULL, ('{"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9553"}}}')::jsonb);
-- router
INSERT INTO configs VALUES (3, 3, NULL, 't', NULL, NULL, ('{"database": {' ||
 '"port": "' || current_setting('migration.db_port') || '", ' ||
  '"schema": "admin_service", ' ||
   '"address": "' || current_setting('migration.db_host') || '", ' ||
    '"database": "' || current_setting('migration.db_dbname') || '", ' ||
     '"password": "' || current_setting('migration.db_pass') || '", ' ||
      '"username": "' || current_setting('migration.db_user') || '", ' ||
       '"createSchema": true}, ' ||
        '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9554"}}' ||
         '}')::jsonb);
-- system
INSERT INTO configs VALUES (4, 4, NULL, 't', NULL, NULL, ('{"dB": {' ||
 '"port": "' || current_setting('migration.db_port') || '", ' ||
  '"schema": "system_service", ' ||
   '"address": "' || current_setting('migration.db_host') || '", ' ||
    '"database": "' || current_setting('migration.db_dbname') || '", ' ||
     '"password": "' || current_setting('migration.db_pass') || '", ' ||
      '"username": "' || current_setting('migration.db_user') || '", ' ||
       '"createSchema": true}, "redisAddress": {' ||
        '"ip": "' || current_setting('migration.redis_host') || '", ' ||
         '"port": "' || current_setting('migration.redis_port') || '"' ||
          '}, ' ||
           '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9555"}}' ||
            '}')::jsonb);
-- admin
INSERT INTO configs VALUES (5, 5, NULL, 't', NULL, NULL, ('{"database": {' ||
 '"port": "' || current_setting('migration.db_port') || '", ' ||
  '"schema": "admin_service", ' ||
   '"address": "' || current_setting('migration.db_host') || '", ' ||
    '"database": "' || current_setting('migration.db_dbname') || '", ' ||
     '"password": "' || current_setting('migration.db_pass') || '", ' ||
      '"username": "' || current_setting('migration.db_user') || '", ' ||
       '"createSchema": true}, ' ||
        '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9557"}}' ||
         '}')::jsonb);

SELECT setval(pg_get_serial_sequence('configs', 'id'), 6, FALSE);

-- +goose Down
DELETE FROM configs WHERE id IN (1,2,3,4,5)
