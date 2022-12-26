-- +goose Up
-- +goose StatementBegin
SET TIMEZONE = 'UTC';
CREATE SCHEMA IF NOT EXISTS homepage_schema;
SET search_path TO homepage_schema,public;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path TO public;
DROP SCHEMA IF EXISTS homepage_schema CASCADE;
-- +goose StatementEnd
