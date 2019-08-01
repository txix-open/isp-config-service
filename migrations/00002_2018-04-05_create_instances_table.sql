-- +goose Up
CREATE TABLE instances (
    id serial4 NOT NULL PRIMARY KEY,
    uuid UUID NOT NULL,
    "name" varchar(255) NOT NULL,
    created_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_created_modified_column_date()
    RETURNS TRIGGER AS
$body$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        NEW.created_at = OLD.created_at;
        NEW.updated_at = (now() at time zone 'utc');
    ELSIF TG_OP = 'INSERT' THEN
        NEW.updated_at = (now() at time zone 'utc');
    END IF;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_created_column_date()
    RETURNS TRIGGER AS
$body$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        NEW.created_at = OLD.created_at;
    ELSIF TG_OP = 'INSERT' THEN
        NEW.created_at = (now() at time zone 'utc');
    END IF;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_customer_modtime BEFORE INSERT OR UPDATE ON instances
    FOR EACH ROW EXECUTE PROCEDURE update_created_column_date();

ALTER TABLE instances ADD CONSTRAINT "UQ_instances_uuid" UNIQUE ("uuid");


-- +goose Down
DROP TABLE instances;
