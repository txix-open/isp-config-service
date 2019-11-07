-- +goose Up

DROP TABLE instances CASCADE;

DROP TRIGGER update_config_create_time ON configs;
DROP TRIGGER deactivate_config on configs;
DROP TRIGGER update_schema_create_time ON config_schemas;
DROP TRIGGER update_customer_modtime ON modules;


ALTER TABLE configs
    DROP CONSTRAINT "FK_configs_moduleId_modules_id";

ALTER TABLE config_schemas
    DROP CONSTRAINT "fk_module_id__module_id";

ALTER TABLE configs
    ALTER COLUMN id TYPE varchar(255),
    ALTER COLUMN module_id TYPE varchar(255),
    ADD COLUMN common_configs varchar(255)[] NOT NULL DEFAULT '{}';

ALTER TABLE config_schemas
    ALTER COLUMN module_id TYPE varchar(255),
    ALTER COLUMN id TYPE varchar(255);

ALTER TABLE modules
    ALTER COLUMN id TYPE varchar(255),
    DROP COLUMN active,
    DROP COLUMN instance_id;


CREATE TABLE common_configs
(
    id          VARCHAR(255) NOT NULL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL DEFAULT 'unnamed',
    description TEXT,
    created_at  timestamp    NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at  timestamp    NOT NULL DEFAULT (now() at time zone 'utc'),
    data        jsonb        NOT NULL DEFAULT '{}'
);


-- +goose Down

DROP TABLE common_configs;

