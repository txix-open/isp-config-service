-- +goose Up
set migration.nats_host = '10.250.9.117';
set migration.nats_port = '4223';
set migration.nats_cluster_id = 'test-cluster';

set migration.rabbit_host = '10.250.9.117';
set migration.rabbit_port = '5672';
set migration.rabbit_user = 'guest';
set migration.rabbit_password = 'guest';


WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-notify', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{
  "isNatsLoggerEnabled": false,
  "nats": {
    "address": {
      "port": "' || current_setting('migration.nats_port') || '",
      "ip": "' || current_setting('migration.nats_host') || '"
    },
    "clusterId": "' || current_setting('migration.nats_cluster_id') || '"
  },
  "queue": {
    "init": {
      "concurrentConsumers": 8,
      "name": "init",
      "prefetchCount": 10
    },
    "tempQueue": {
      "name": "tempt",
      "prefetchCount": 10,
      "concurrentConsumers": 128
    },
    "defaultQueue": {
      "concurrentConsumers": 128,
      "name": "events",
      "prefetchCount": 10
    }
  },
  "rabbit": {
    "password": "' || current_setting('migration.rabbit_password') || '",
    "reconnectionTimeoutMs": 0,
    "user": "' || current_setting('migration.rabbit_user') || '",
    "address": {
      "ip": "' || current_setting('migration.rabbit_host') || '",
      "port": "' || current_setting('migration.rabbit_port') || '"
    }
  },
  "metrics": {
    "memory": true,
    "address": {
      "path": "/metrics",
      "port": "9569",
      "ip": "0.0.0.0"
    },
    "gc": true
  }
}')::jsonb);

-- +goose Down

