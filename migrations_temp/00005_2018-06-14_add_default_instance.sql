-- +goose Up
INSERT INTO instances VALUES (1, 'bf482806-0c3d-4e0d-b9d4-12c037b12d70', 'main', NULL);
SELECT setval(pg_get_serial_sequence('instances', 'id'), 2, FALSE);

-- +goose Down
DELETE FROM instances WHERE id=1
