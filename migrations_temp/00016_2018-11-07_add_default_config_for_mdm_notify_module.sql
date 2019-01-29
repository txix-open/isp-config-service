-- +goose Up
set migration.nats_host = 'isp-nats';
set migration.nats_port = '4223';
set migration.nats_cluster_id = 'test-cluster';

set migration.rabbit_host = 'isp-rabbit';
set migration.rabbit_port = '5672';
set migration.rabbit_user = 'guest';
set migration.rabbit_password = 'guest';
set migration.rabbit_system_queue_prefix = 'isp-test';

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-notify', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{
    "notifyConfigurations":[],
    "logRabbitEnabled": false,
    "logger": {
      "asyncLogger":{
        "enable": false,
        "nats": {
          "address": {
            "port": "' || current_setting('migration.nats_port') || '",
            "ip": "' || current_setting('migration.nats_host') || '"
          },
          "clusterId": "' || current_setting('migration.nats_cluster_id') || '"
        }
      },
      "syncLogger": {
        "enable": false,
        "filename": "/opt/isp/isp-mdm-notify-service/events_logs/events.log"
      },
      "sendAddRequestSudirV1IsOn": true,
      "sendFindRequestSudirV1IsOn": true,
      "sendUpdateRequestSudirV1IsOn": true,
      "sendUpdateRequestSudirV2IsOn": true,
      "creatingSystemNotificationsIsOn": true,
      "sendErlRequestQueueIsOn": true,
      "sendJsonRequestQueueIsOn": true
    },
    "metrics":{
      "gc": true,
      "memory": true,
      "address": {
        "port": "9572",
        "ip": "0.0.0.0",
        "path": "/metrics"
      }
    },
    "rabbit": {
      "routingKey": "notification",
      "exchangeName": "notifications",
      "systemQueuesPrefix": "' || current_setting('migration.rabbit_system_queue_prefix') || '",
      "declareQueueAndExchange": true,
      "password": "' || current_setting('migration.rabbit_password') || '",
      "user": "' || current_setting('migration.rabbit_user') || '",
      "address": {
        "ip": "' || current_setting('migration.rabbit_host') || '",
        "port": "' || current_setting('migration.rabbit_port') || '"
      },
      "queue": {
        "name": "notifier.queue",
        "prefetchCount": 1000
      }
    }
  }')::jsonb);

-- +goose Down

