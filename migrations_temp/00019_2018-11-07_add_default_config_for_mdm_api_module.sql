-- +goose Up
WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-api', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{' ||
   '"pKAttributeName": "id"' ||
    '}')::jsonb);

-- +goose Down

