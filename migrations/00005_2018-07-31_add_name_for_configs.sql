-- +goose Up
ALTER TABLE configs
  ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT 'unnamed',
  ADD COLUMN description TEXT;

-- +goose Down
ALTER TABLE configs
  DROP COLUMN name,
  DROP COLUMN description;
