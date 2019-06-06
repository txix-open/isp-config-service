-- +goose Up
set migration.db_host = 'isp-pgsql';
set migration.db_port = '5432';
set migration.db_dbname = 'isp-test';
set migration.db_user = 'isp-test';
set migration.db_pass = '123321';

set migration.redis_host = 'isp-redis';
set migration.redis_port = '5672';

set migration.admin_token_secret = 'secret';

-- auth
INSERT INTO configs VALUES (1, 1, NULL, 't', NULL, NULL, (format('{
  "jwt": {
    "secret": "%s"
  },
  "redis": {
    "ip": "%s",
    "port": %s
  },
  "header": {
    "token": {
      "user": "X-USER-TOKEN",
      "admin": "X-AUTH-ADMIN",
      "device": "X-DEVICE-TOKEN",
      "application": "X-APPLICATION-TOKEN"
    },
    "identity": {
      "device": "X-DEVICE-IDENTITY",
      "domain": "X-DOMAIN-IDENTITY",
      "system": "X-SYSTEM-IDENTITY",
      "service": "X-SERVICE-IDENTITY",
      "application": "X-APPLICATION-IDENTITY",
      "user": "X-USER-IDENTITY"
    }
  },
  "unauthorized.html": "<HTML><BODY><H1>401 Unauthorized BAD CREDENTIALS</H1></BODY></HTML>"
}',
    current_setting('migration.admin_token_secret'),
    current_setting('migration.redis_host'),
    current_setting('migration.redis_port')
))::jsonb);
-- converter
INSERT INTO configs VALUES (2, 2, NULL, 't', NULL, NULL, ('{"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9553"}}}')::jsonb);
-- router
INSERT INTO configs VALUES (3, 3, NULL, 't', NULL, NULL, ('{
  "metrics": {
    "gc": true,
    "memory": true,
    "address": {
      "port": "9554",
      "ip": "0.0.0.0",
      "path": "/metrics"
    }
  }
}')::jsonb);
-- system
INSERT INTO configs VALUES (4, 4, NULL, 't', NULL, NULL, (format('{
  "dB": {
    "password": "%s",
    "username": "%s",
    "createSchema": true,
    "port": "%s",
    "schema": "system_service",
    "address": "%s",
    "database": "%s"
  },
  "metrics": {
    "address": {
      "ip": "0.0.0.0",
      "path": "/metrics",
      "port": "9555"
    },
    "gc": true,
    "memory": true
  },
  "redisAddress": {
    "ip": "%s",
    "port": "%s"
  }
}', current_setting('migration.db_pass'),
    current_setting('migration.db_user'),
    current_setting('migration.db_port'),
    current_setting('migration.db_host'),
    current_setting('migration.db_dbname'),
    current_setting('migration.redis_host'),
    current_setting('migration.redis_port')
))::jsonb);
-- admin
INSERT INTO configs VALUES (5, 5, NULL, 't', NULL, NULL, (format('{
  "metrics": {
    "gc": true,
    "memory": true,
    "address": {
      "path": "/metrics",
      "port": "9557",
      "ip": "0.0.0.0"
    }
  },
  "database": {
    "password": "%s",
    "username": "%s",
    "createSchema": true,
    "port": "%s",
    "schema": "admin_service",
    "address": "%s",
    "database": "%s"
  },
  "secretKey": "%s"
}',
    current_setting('migration.db_pass'),
    current_setting('migration.db_user'),
    current_setting('migration.db_port'),
    current_setting('migration.db_host'),
    current_setting('migration.db_dbname'),
    current_setting('migration.admin_token_secret')
))::jsonb);

SELECT setval(pg_get_serial_sequence('configs', 'id'), 6, FALSE);

-- +goose Down
DELETE FROM configs WHERE id IN (1,2,3,4,5)

