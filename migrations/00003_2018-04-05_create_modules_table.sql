-- +goose Up
CREATE TABLE modules (
    id serial4 NOT NULL PRIMARY KEY,
    instance_id int4 NOT NULL,
    "name" varchar(255) NOT NULL,
    "active" bool DEFAULT true,
    created_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL,
    last_connected_at timestamp,
    last_disconnected_at timestamp
);

ALTER TABLE modules
    ADD CONSTRAINT "FK_modules_instanceId_instances_id"
    FOREIGN KEY ("instance_id") REFERENCES instances ("id")
    ON DELETE CASCADE ON UPDATE CASCADE;
CREATE INDEX IX_modules_instanceId ON modules USING hash (instance_id);

CREATE TRIGGER update_customer_modtime BEFORE INSERT OR UPDATE ON modules
    FOR EACH ROW EXECUTE PROCEDURE update_created_column_date();

ALTER TABLE modules ADD CONSTRAINT "UQ_modules_instanceId_name" UNIQUE ("instance_id", "name");

-- +goose Down
DROP TABLE modules;
