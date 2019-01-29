-- +goose Up
INSERT INTO modules VALUES (1, 1, 'auth', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (2, 1, 'converter', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (3, 1, 'router', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (4, 1, 'system', 't', NULL, NULL, NULL);
INSERT INTO modules VALUES (5, 1, 'admin', 't', NULL, NULL, NULL);
SELECT setval(pg_get_serial_sequence('modules', 'id'), 6, FALSE);

-- +goose Down
DELETE FROM modules WHERE id IN (1,2,3,4,5);
