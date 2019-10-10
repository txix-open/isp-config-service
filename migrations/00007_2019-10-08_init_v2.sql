-- +goose Up

DROP TABLE instances CASCADE;

ALTER TABLE configs
    ADD COLUMN uuid           UUID   NOT NULL DEFAULT uuid_generate_v4(),
    ADD COLUMN common_configs UUID[] NOT NULL DEFAULT '{}',
    ADD CONSTRAINT "unique_configs_uuid" UNIQUE ("uuid");

ALTER TABLE config_schemas
    ADD COLUMN uuid UUID NOT NULL DEFAULT uuid_generate_v4(),
    ADD CONSTRAINT "unique_config_schemas_uuid" UNIQUE ("uuid");

ALTER TABLE modules
    ADD COLUMN uuid UUID NOT NULL DEFAULT uuid_generate_v4(),
    ADD CONSTRAINT "unique_modules_uuid" UNIQUE ("uuid");


CREATE TABLE common_configs
(
    id          serial4      NOT NULL PRIMARY KEY,
    uuid        UUID         NOT NULL,
    name        VARCHAR(255) NOT NULL DEFAULT 'unnamed',
    description TEXT,
    version     int4         NOT NULL,
    active      bool         NOT NULL DEFAULT false,
    created_at  timestamp    NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at  timestamp    NOT NULL DEFAULT (now() at time zone 'utc'),
    data        jsonb        NOT NULL DEFAULT '{}'
);

ALTER TABLE common_configs
    ADD CONSTRAINT "unique_common_configs_uuid" UNIQUE ("uuid");


-- +goose Down

ALTER TABLE configs
    DROP COLUMN uuid;

ALTER TABLE config_schemas
    DROP COLUMN uuid;

ALTER TABLE modules
    DROP COLUMN uuid;

DROP TABLE common_configs;

