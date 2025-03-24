-- +goose Up
-- +goose NO TRANSACTION
create table isp_config_service__variable
(
    name        text primary key,
    description text    not null,
    type        text    not null,
    value       text    not null,
    created_at  integer not null default (unixepoch()),
    updated_at   integer not null default (unixepoch())
);

create table isp_config_service__config_has_variable
(
    config_id     text    not null,
    variable_name text    not null,
    primary key (config_id, variable_name),
    foreign key (config_id) references isp_config_service__config (id) on delete cascade on update cascade,
    foreign key (variable_name) references isp_config_service__variable (name)
);

create index IX_isp_config_service__config_has_variable__variable_name on isp_config_service__config_has_variable (variable_name);

-- +goose Down
drop table isp_config_service__config_has_variable;
drop table isp_config_service__variable;