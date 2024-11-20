-- +goose Up
-- +goose NO TRANSACTION
drop table isp_config_service__backend;

create table isp_config_service__backend
(
    ws_connection_id       text    not null,
    module_id              text    not null,
    address                text    not null,
    version                text    not null,
    lib_version            text    not null,
    module_name            text    not null,
    config_service_node_id text    not null,
    endpoints              blob    not null,
    required_modules       blob    not null,
    created_at             integer not null default (unixepoch()),
    primary key (ws_connection_id),
    foreign key (module_id) references isp_config_service__module (id) on delete cascade on update cascade
);

create index IX_isp_config_service__backend__module_id on isp_config_service__backend (module_id);
create index IX_isp_config_service__backend__node_id on isp_config_service__backend(config_service_node_id);

-- +goose Down
