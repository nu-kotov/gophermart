-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id       UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    login         TEXT NOT NULL,
    password      TEXT NOT NULL,
    UNIQUE (login)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
