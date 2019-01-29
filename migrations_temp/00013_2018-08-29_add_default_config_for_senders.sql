-- +goose Up
set migration.nats_host = 'isp-nats';
set migration.nats_port = '4223';
set migration.nats_cluster_id = 'test-cluster';

WITH moduleVar AS (INSERT INTO modules (instance_id, name, active) VALUES (1, 'mobile-push', 't') RETURNING id)
INSERT INTO configs (module_id, active, data)
VALUES ((SELECT id FROM moduleVar), 't', (
    format('{
    "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9561"}},
  "firebaseConfig": {
    "token_uri": "https://accounts.google.com/o/oauth2/token",
    "client_id": "",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "client_email": "firebase-adminsdk-f70qo@test-ccdab.iam.gserviceaccount.com",
    "private_key": "",
    "private_key_id": "",
    "project_id": "test-ccdab",
    "type": "service_account",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-f70qo%%40test-ccdab.iam.gserviceaccount.com"
  },
  "concurrentSendersCount": 6,
  "natsConfig": {
    "address": {
      "port": "%s",
      "ip": "%s"
    },
    "clusterId": "%s"
  }
}', current_setting('migration.nats_port'),
    current_setting('migration.nats_host'),
    current_setting('migration.nats_cluster_id')
)) :: jsonb);

WITH moduleVar AS (INSERT INTO modules (instance_id, name, active) VALUES (1, 'sms', 't') RETURNING id)
INSERT INTO configs (module_id, active, data)
VALUES ((SELECT id FROM moduleVar), 't', (format('{
  "natsConfig": {
    "clusterId": "%s",
    "address": {
      "ip": "%s",
      "port": "%s"
    }
  },
  "smppClients": [
    {
      "default": true,
      "poolSize": 1,
      "connection": {
        "password": "",
        "user": "isp",
        "address": {
          "ip": "5.187.0.65",
          "port": "9931"
        }
      },
      "ResponseTimeoutMs": 5000,
      "concurrentSendersCount": 2,
      "name": "smsc"
    }
  ],
  "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9565"}}
}', current_setting('migration.nats_cluster_id'),
    current_setting('migration.nats_host'),
    current_setting('migration.nats_port')
)) :: jsonb);

WITH moduleVar AS (INSERT INTO modules (instance_id, name, active) VALUES (1, 'email', 't') RETURNING id)
INSERT INTO configs (module_id, active, data)
VALUES ((SELECT id FROM moduleVar), 't', (format('{
  "natsConfig": {
    "clusterId": "%s",
    "address": {
      "ip": "%s",
      "port": "%s"
    }
  },
  "smtpClients": [

  ],
  "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9558"}}
}', current_setting('migration.nats_cluster_id'),
    current_setting('migration.nats_host'),
    current_setting('migration.nats_port')
)) :: jsonb);

-- +goose Down
