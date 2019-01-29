-- +goose Up
WITH moduleVar AS (INSERT INTO modules (instance_id, name, active) VALUES (1, 'file-storage', 't')
RETURNING id)
INSERT INTO configs (module_id, active, data)
VALUES ((SELECT id FROM moduleVar), 't', '{
  "imageSizeList": [
    75,
    130,
    200,
    320,
    510,
    604,
    807
  ],
  "rootDirectory": "./files/",
  "maxRequestBodySizeInBytes": 1073741824,
  "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9559"}}
}' :: jsonb);

WITH moduleVar AS (INSERT INTO modules (instance_id, name, active) VALUES (1, 'rest-backend', 't')
RETURNING id)
INSERT INTO configs (module_id, active, data)
VALUES ((SELECT id FROM moduleVar), 't', '{
  "backends": [], "metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9564"}}
}' :: jsonb);

-- +goose Down
