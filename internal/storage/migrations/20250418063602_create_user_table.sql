-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id       uuid NOT NULL PRIMARY KEY,
    user_login    TEXT NOT NULL,
    user_password TEXT NOT NULL,
    UNIQUE (user_login)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
