-- +goose Up
-- +goose NO TRANSACTION
create table isp_config_service__module
(
    id                   text    not null primary key,
    name                 text    not null unique,
    last_connected_at    integer,
    last_disconnected_at integer,
    created_at           integer not null default (unixepoch())
);

create table isp_config_service__config
(
    id         text    not null primary key,
    name       text    not null,
    module_id  text    not null,
    data       blob    not null,
    version    int     not null default 1,
    active     int     not null default 0,
    admin_id   int     not null default 0,
    created_at integer not null default (unixepoch()),
    updated_at integer not null default (unixepoch()),
    foreign key (module_id) references isp_config_service__module (id) on delete cascade on update cascade
);

create index IX_isp_config_service__config__module_id on isp_config_service__config (module_id);

create table isp_config_service__config_history
(
    id         text    not null primary key,
    config_id  text    not null,
    data       blob    not null,
    version    int     not null default 1,
    admin_id   int     not null default 0,
    created_at integer not null default (unixepoch()),
    foreign key (config_id) references isp_config_service__config (id) on delete cascade on update cascade
);

create index IX_isp_config_service__config_history__config_id on isp_config_service__config_history (config_id);

create table isp_config_service__config_schema
(
    id             text    not null primary key,
    module_id      text    not null unique,
    data           blob    not null,
    module_version text    not null default 1,
    created_at     integer not null default (unixepoch()),
    updated_at     integer not null default (unixepoch()),
    foreign key (module_id) references isp_config_service__module (id) on delete cascade on update cascade
);

create table isp_config_service__backend
(
    module_id        text    not null,
    address          text    not null,
    version          text    not null,
    lib_version      text    not null,
    module_name      text    not null,
    endpoints        blob    not null,
    required_modules blob    not null,
    created_at       integer not null default (unixepoch()),
    updated_at       integer not null default (unixepoch()),
    primary key (module_id, address),
    foreign key (module_id) references isp_config_service__module (id) on delete cascade on update cascade
);

create table isp_config_service__event
(
    id         integer,
    payload    blob    not null,
    created_at integer not null default (unixepoch()),
    primary key (id)
);

-- +goose Down
drop table isp_config_service__module;
drop table isp_config_service__config;
drop table isp_config_service__config_history;
drop table isp_config_service__config_schema;
drop table isp_config_service__backend;
drop table isp_config_service__event;
