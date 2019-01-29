-- +goose Up
INSERT INTO modules VALUES (6, 1, 'user', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (7, 1, 'mobile', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (8, 1, 'awip', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (9, 1, 'mongo-backend', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (10, 1, 'sql-backend', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (11, 1, 'events', 't', NULL, NULL, NULL);
SELECT setval(pg_get_serial_sequence('modules', 'id'), 12, FALSE);

-- +goose Down
DELETE FROM modules WHERE id IN (6,7,8,9,10,11);
