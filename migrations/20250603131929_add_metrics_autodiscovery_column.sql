-- +goose Up
-- +goose NO TRANSACTION
alter table isp_config_service__backend add column metrics_autodiscovery blob;

-- +goose Down
alter table isp_config_service__backend drop column metrics_autodiscovery;
