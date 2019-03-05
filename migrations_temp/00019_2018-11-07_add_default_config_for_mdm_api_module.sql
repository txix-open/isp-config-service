-- +goose Up
WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-api', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{
   "metrics": {
       "memory": true,
       "address": {
         "path": "/metrics",
         "port": "9570",
         "ip": "0.0.0.0"
       },
       "gc": true
     },
  "outputQueue": {
    "routingKey": "external",
    "enable": false,
    "exchange": "external"
  },
  "appConfigurations": [],
  "limits": {
      "defaultRecordsLimit": 1000,
      "defaultPackageSizeLimit": 1000,
      "defaultAsyncWorkersLimit": 2
  },
  "syncLogger": {
      "enable": false,
      "filename": "/opt/isp/isp-mdm-api-service/events/events.log.gz"
    }
  }')::jsonb);

-- +goose Down

