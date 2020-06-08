-- +goose Up
CREATE TABLE version_config
(
    id             VARCHAR(255) NOT NULL PRIMARY KEY,
    config_id      VARCHAR(255) NOT NULL,
    config_version INT4         NOT NULL,
    data           JSONB        NOT NULL DEFAULT '{}',
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE version_config;

