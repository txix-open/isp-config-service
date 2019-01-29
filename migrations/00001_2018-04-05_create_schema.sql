-- +goose Up
CREATE schema if NOT EXISTS "config_service";

-- +goose Down
DROP schema if EXISTS config_service CASCADE;