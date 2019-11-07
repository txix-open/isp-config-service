-- +goose Up
CREATE TABLE config_schemas (
  id         SERIAL4 PRIMARY KEY,
  module_id  INTEGER      NOT NULL,
  version    VARCHAR(255) NOT NULL,
  schema     JSONB        NOT NULL DEFAULT '{}',
  created_at TIMESTAMP    NOT NULL DEFAULT (now() at time zone 'utc'),
  updated_at TIMESTAMP    NOT NULL DEFAULT (now() at time zone 'utc'),
  CONSTRAINT FK_module_id__module_id FOREIGN KEY (module_id)
  REFERENCES modules (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TRIGGER update_schema_create_time
  BEFORE INSERT OR UPDATE
  ON config_schemas
  FOR EACH ROW EXECUTE PROCEDURE update_created_modified_column_date();

CREATE INDEX IX_schemes_moduleId
  ON config_schemas
  USING hash (module_id);

-- +goose Down
DROP TABLE config_schemas;
