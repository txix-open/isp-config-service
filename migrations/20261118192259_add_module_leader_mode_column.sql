-- +goose Up
-- +goose NO TRANSACTION
alter table isp_config_service__backend add column module_leader_mode text;

-- +goose Down
alter table isp_config_service__backend drop column module_leader_mode;
