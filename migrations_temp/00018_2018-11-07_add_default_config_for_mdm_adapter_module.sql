-- +goose Up
set migration.nats_host = 'isp-nats';
set migration.nats_port = '4223';
set migration.nats_cluster_id = 'test-cluster';

set migration.rabbit_host = 'isp-rabbit';
set migration.rabbit_port = '5672';
set migration.rabbit_user = 'guest';
set migration.rabbit_password = 'guest';

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-adapter', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{
  "currentMode": "default"
  "syncLogger": {
    "filename": "/opt/isp/isp-mdm-adapter-service/events/events.log",
    "enable": true
  },
  "asyncLogger": {
    "enable": false,
    "nats": {
      "clusterId": "' || current_setting('migration.nats_cluster_id') || '",
      "address": {
        "ip": "' || current_setting('migration.nats_host') || '",
        "port": "' || current_setting('migration.nats_port') || '"
      }
    }
  },
  "queue": {
    "init": {
      "concurrentConsumers": 1,
      "name": "init",
      "prefetchCount": 1000
    },
    "tempQueue": {
      "name": "tempt",
      "prefetchCount": 1000,
      "concurrentConsumers": 1
    },
    "defaultQueue": {
      "concurrentConsumers": 1,
      "name": "events",
      "prefetchCount": 1000
    }
  },
  "rabbit": {
    "password": "' || current_setting('migration.rabbit_password') || '",
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
  },
  "outputQueue": {
    "routingKey": "external",
    "enable": false,
    "rabbit": {
      "password": "' || current_setting('migration.rabbit_password') || '",
      "user": "' || current_setting('migration.rabbit_user') || '",
      "address": {
        "ip": "' || current_setting('migration.rabbit_host') || '",
        "port": "' || current_setting('migration.rabbit_port') || '"
      }
    },
    "exchange": "external"
  }
}')::jsonb);

-- +goose Down

