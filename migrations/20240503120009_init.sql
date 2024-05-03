-- +goose Up
-- +goose NO TRANSACTION
create table isp_config_service__module
(
    id                   text not null primary key,
    name                 text not null unique,
    last_connected_at    text,
    last_disconnected_at text,
    created_at           text not null default (datetime('now'))
);

create table isp_config_service__config
(
    id         text not null primary key,
    name       text not null,
    module_id  text not null,
    data       blob not null,
    version    int  not null default 1,
    active     int  not null default 0,
    created_at text not null default (datetime('now')),
    updated_at text not null default (datetime('now')),
    foreign key (module_id) references isp_config_service__module (id) on delete cascade on update cascade
);

create index IX_isp_config_service__config__module_id on isp_config_service__config (module_id);

create table isp_config_service__config_history
(
    id         text not null primary key,
    config_id  text not null,
    data       blob not null,
    version    int  not null default 1,
    admin_id   int  not null,
    created_at text not null default (datetime('now')),
    foreign key (config_id) references isp_config_service__config (id) on delete cascade on update cascade
);

create index IX_isp_config_service__config_history__config_id on isp_config_service__config_history (config_id);

create table isp_config_service__config_schema
(
    id         text not null primary key,
    name       text not null,
    module_id  text not null,
    data       blob not null,
    version    text not null default 1,
    created_at text not null default (datetime('now')),
    updated_at text not null default (datetime('now')),
    foreign key (module_id) references isp_config_service__module (id) on delete cascade on update cascade
);

create index IX_isp_config_service__config_schema__module_id on isp_config_service__config_schema (module_id);

-- +goose Down
drop table isp_config_service__module;
