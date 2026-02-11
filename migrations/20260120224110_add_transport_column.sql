-- +goose Up
-- +goose NO TRANSACTION
alter table isp_config_service__backend add column transport text;

-- +goose Down
alter table isp_config_service__backend drop column transport;
