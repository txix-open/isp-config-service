-- +goose Up

DROP TABLE instances CASCADE;

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
    ALTER COLUMN id TYPE varchar(255);

-- Constraints

ALTER TABLE configs
    ADD CONSTRAINT "FK_configs_moduleId_modules_id"
        FOREIGN KEY ("module_id") REFERENCES modules ("id")
            ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE config_schemas
    ADD CONSTRAINT FK_module_id__module_id FOREIGN KEY (module_id)
        REFERENCES modules (id) ON UPDATE CASCADE ON DELETE CASCADE;



CREATE TABLE common_configs
(
    id          VARCHAR(255) NOT NULL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL DEFAULT 'unnamed',
    description TEXT,
    version     int4         NOT NULL,
    active      bool         NOT NULL DEFAULT false,
    created_at  timestamp    NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at  timestamp    NOT NULL DEFAULT (now() at time zone 'utc'),
    data        jsonb        NOT NULL DEFAULT '{}'
);


-- +goose Down

DROP TABLE common_configs;

