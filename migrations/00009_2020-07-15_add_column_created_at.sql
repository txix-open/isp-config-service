-- +goose Up
ALTER TABLE version_config
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT (now() at time zone 'utc');

-- +goose Down


