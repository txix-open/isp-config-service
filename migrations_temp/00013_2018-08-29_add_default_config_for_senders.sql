-- +goose Up
set migration.nats_host = '10.250.9.117';
set migration.nats_port = '4223';
set migration.nats_cluster_id = 'test-cluster';

WITH moduleVar AS (INSERT INTO modules (instance_id, name, active) VALUES (1, 'mobile-push', 't') RETURNING id)
INSERT INTO configs (module_id, active, data)
VALUES ((SELECT id FROM moduleVar), 't', (
    format('{
    "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9561"}},
  "firebaseConfig": {
    "token_uri": "https://accounts.google.com/o/oauth2/token",
    "client_id": "115553569331708875508",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "client_email": "firebase-adminsdk-f70qo@test-ccdab.iam.gserviceaccount.com",
    "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCM7Rl7AmGQSOxx\nUdQs+yY7cwaUmqi2khmsre0/kfJU2zCJxotbYNeyfWJZk3D8hngxnmI6hrYqWEq3\nyy/J8rhI7AmcY3eXC1yyL/sKTClFjyZVEoDcBWTn0HOuM8YAQAqRAx2vznKDquoj\nLgUHl09qTxsVRmPdfc7fKE45OZBdetA/kdIEd4S5J+CKeYB7tpuT+0eL4QUNOzE1\n6SGqYnhA97QD7K3ylhWxKC8U0Ozyqub/OTg7WRLEYV36Z+6Ok0BZkYiCXbUNukST\nXaKc8DGAFrzcu7jW8rJIRZ/uc/Im5judjuMRznzu1SMPVT6pL8HfduGFc23FIAL/\nGjmcDwWbAgMBAAECggEAETe6oDvHPcCbGrE7sg8xOZwxFqDasguhlWZekSC8sb9h\n68NVLWHkmIsXJAiOilvHfZBzQeFJilzlLBVoDk1YVJh6CCBi8RJTTfXsvvJVLIlz\nznsHQVprXKMsLwFmVIt+fv8ZdmxLs2iDWK77sFS9QCjQD0ZdVydSyhL7k6RDzhfP\nw6It0H3JyyqD8W7DOvYSUfi/oZKZvufBAkEauhuiVVyite97cJtV8YNmiVCznOtK\nOw0OpuzemKez0rh673VtEEvUgrvOy5WNtjunoRkNKZkW1JiAWYpf6FpOOFub/lG3\nQPFwjLZeEUa8xc/E2JJP6HQMON0qF5zvtKr4QH/9uQKBgQDByXfLh+56Cu+6E0CL\nigH0GmagwZQnw7/+zhi1AhC2DjxAmXq9Y0kFrmpwr6FXqtRJm5AANOipASsRknaq\n5kAlDCGdqHnw1LDlNbrxywDvTYADo13ebx2cXfxESDfx0gofXehGhfFfOyXJ0lwp\ngXbgxIQFNjv2QWjSdb637ZaOrQKBgQC6KzmZcMAWo7j+eGo0WoAVTCVu/SDcV/sG\nF4+NxhLjMYSuMbBF7BdMy1rJ6g3SrdULGzw0KDeEqo1XzneOyzHOnEiZFKTkgIPQ\n347fWOYTBKMNEJd37jAIHKrrFiyD/pvL2ZLu1mSnCHsi/v5u02gqpf8Vkvwo/5ME\n4BuhB/PWZwKBgQC7V0adl+rfR0VcURJcE+4xi3hdvua4zpAFCD9wde+r4PU0ymuT\nPbGxcV1rVQ8YTojuJBrBaGToRb3aPgrEytWGO0UgQmiofyYIYLo62LMtpXG1krDD\nwg4RRfcEGAEloZWxnzpXO1QOaYLtqpT4dzVys+ihlT3AoplwpO3cqC6d/QKBgEGu\nkmnSX9Mc/F27eizyaRIahXJ9GCTlXYkustUgNvW1OMyEd16UBzxu2p82Vp4n+mwq\ntbjpH31M9wUtsPzOL8pnVS29HNgJh3ggB7ZBFRtMnYI0glwrywJxqtO6RQZkw+7N\n2ostVOGhmmAkevv61luFqVOhQhns4Z/suZK8zYitAoGAHwLWUQViCy0HFkhED5GV\nWrZZFRH2ucpK+ljOp7dbbUPA+dTD2qu/hho3QCaT5odJWap42TbSL6YhZg/7jN7h\nOSQW8DiDPzjUtL1gfqOxJdQavrgR2a2AW054loM0R508699Vx6Hir/ayD4nZ9IUo\nm2meeJVZcknz/L9LXOOIUBk=\n-----END PRIVATE KEY-----\n",
    "private_key_id": "88d1af761304de758388747d72931c3457df54ef",
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
        "password": "1q2w3e!",
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
