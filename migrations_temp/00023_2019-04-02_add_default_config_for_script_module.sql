-- +goose Up

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'script', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{
      "scripts": [

      ],
      "shaderScript": "",
      "scriptExecutionTimeoutMs": 1000
    }')::jsonb
  );

-- +goose Down

