-- +goose Up
set migration.db_host = 'isp-pgsql';
set migration.db_port = '5432';
set migration.db_dbname = 'isp-test';
set migration.db_user = 'isp-test';
set migration.db_pass = '123321';

set migration.redis_host = 'isp-redis';
set migration.redis_port = '6379';

-- user
INSERT INTO configs VALUES (6, 6, NULL, 't', NULL, NULL, (format('{
  "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9556"}},
  "database": {
    "username": "%s",
    "createSchema": true,
    "port": "%s",
    "schema": "user_service",
    "address": "%s",
    "database": "%s",
    "password": "%s"
  },
  "redisAddress": {
    "ip": "%s",
    "port": "%s"
  },
  "userTokenExpireMs": -1,
  "authSmsSystemSender": "",
  "authSmsTextTemplate": "Your code {{code}}",
  "deviceTokenExpireMs": -1,
  "authSmsShortCodeLength": 4,
  "authSmsShortCodeExpireMs": 300000
}', current_setting('migration.db_user'),
    current_setting('migration.db_port'),
    current_setting('migration.db_host'),
    current_setting('migration.db_dbname'),
    current_setting('migration.db_pass'),
    current_setting('migration.redis_host'),
    current_setting('migration.redis_port')
)) :: jsonb);
-- mobile
INSERT INTO configs VALUES (7, 7, NULL, 't', NULL, NULL, ('{"database": {' ||
 '"port": "' || current_setting('migration.db_port') || '", ' ||
  '"schema": "mobile_service", ' ||
   '"address": "' || current_setting('migration.db_host') || '", ' ||
    '"database": "' || current_setting('migration.db_dbname') || '", ' ||
     '"password": "' || current_setting('migration.db_pass') || '", ' ||
      '"username": "' || current_setting('migration.db_user') || '", ' ||
       '"createSchema": true}, ' ||
        '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9562"}}' ||
         '}')::jsonb);
-- awip
INSERT INTO configs VALUES (8, 8, NULL, 't', NULL, NULL, '{
    "backends": [
      {
        "id": "450e754f-aafb-4a9c-8d87-23c6a2da9833",
        "url": "http://{{ cname }}-nginx.{{ main_domain }}:18888/ru.isp.starter.awip.backend.ws/backend?wsdl",
        "name": "citymatica_user_new",
        "wrapList": true,
        "createDate": "2018-09-07T09:41:03.853Z",
        "clientPoolSize": 0,
        "requestTimeout": 30000,
        "systemIdentity": "1"
      },
      {
        "id": "74d76314-b28b-11e8-96f8-529269fb1459",
        "url": "http://{{ cname }}-nginx.{{ main_domain }}:18880/ru.mp.storage.backend.ws/backend?wsdl",
        "name": "citymatica_user_old",
        "wrapList": true,
        "createDate": "2018-09-07T09:41:03.853Z",
        "clientPoolSize": 0,
        "requestTimeout": 30000,
        "systemIdentity": "1"
      },
      {
        "id": "74d7659e-b28b-11e8-96f8-529269fb1459",
        "url": "http://{{ cname }}-nginx.{{ main_domain }}:18890/ru.isp.starter.awip.backend.ws/backend?wsdl",
        "name": "citymatica_user_object_storage",
        "wrapList": true,
        "createDate": "2018-09-07T09:41:03.853Z",
        "clientPoolSize": 0,
        "requestTimeout": 30000,
        "systemIdentity": "1"
      },
      {
        "id": "74d75cc0-b28b-11e8-96f8-529269fb1459",
        "url": "http://{{ cname }}-nginx.{{ main_domain }}:18891/ru.isp.starter.awip.backend.ws/backend?wsdl",
        "name": "citymatica_user_events",
        "wrapList": true,
        "createDate": "2018-09-07T09:41:03.853Z",
        "clientPoolSize": 0,
        "requestTimeout": 30000,
        "systemIdentity": "1"
      },
      {
        "id": "d5575047-6588-46d9-be58-b487d42d217c",
        "url": "http://{{ cname }}-nginx.{{ main_domain }}:18889/ru.isp.starter.awip.backend.ws/backend?wsdl",
        "name": "citymatica_operator",
        "wrapList": true,
        "createDate": "2018-09-07T09:41:03.853Z",
        "clientPoolSize": 0,
        "requestTimeout": 30000,
        "systemIdentity": "1"
      },
      {
        "id": "74d76094-b28b-11e8-96f8-529269fb1459",
        "url": "http://{{ cname }}-nginx.{{ main_domain }}:18895/ru.isp.starter.awip.backend.ws/backend?wsdl",
        "name": "citymatica_operator_metrics",
        "wrapList": true,
        "createDate": "2018-09-07T09:41:03.853Z",
        "clientPoolSize": 0,
        "requestTimeout": 30000,
        "systemIdentity": "1"
      }
    ]
}');
-- mongo-backend
INSERT INTO configs VALUES (9, 9, NULL, 't', NULL, NULL, '{"connections": [],"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9563"}}}');
-- sql-backend
INSERT INTO configs VALUES (10, 10, NULL, 't', NULL, NULL, '{"connections": [], "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9566"}}}');
-- events
INSERT INTO configs VALUES (11, 11, NULL, 't', NULL, NULL, '{"eventRoutes": [{"handlers": {}, "applicationGroupId": 1}, {"handlers": {}, "applicationGroupId": 5}], "queueEndpoints": []}');

SELECT setval(pg_get_serial_sequence('configs', 'id'), 12, FALSE);

-- +goose Down
DELETE FROM configs WHERE id IN (6,7,8,9,10,11)
