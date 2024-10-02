-- +goose Up
-- +goose NO TRANSACTION
CREATE TABLE isp_config_service__event_autoincrement (
    id integer primary key autoincrement,
    payload blob not null,
    created_at integer not null default (unixepoch())
);

INSERT INTO
    isp_config_service__event_autoincrement (id, payload, created_at)
SELECT id, payload, created_at
FROM isp_config_service__event;
DROP TABLE isp_config_service__event;

ALTER TABLE isp_config_service__event_autoincrement RENAME TO isp_config_service__event;
-- +goose Down
CREATE TABLE isp_config_service__event_without_autoincrement (
    id integer primary key,
    payload blob not null,
    created_at integer not null default (unixepoch())
);

INSERT INTO
    isp_config_service__event_without_autoincrement (id, payload, created_at)
SELECT id, payload, created_at
FROM isp_config_service__event;

DROP TABLE isp_config_service__event;

ALTER TABLE isp_config_service__event_without_autoincrement
RENAME TO isp_config_service__event;