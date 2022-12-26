-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS homepage_schema.user(
        id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY NOT NULL,
        name TEXT NOT NULL,
        age  INTEGER NOT NULL,
        city TEXT,
        phone TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS homepage_schema.user CASCADE;
-- +goose StatementEnd
