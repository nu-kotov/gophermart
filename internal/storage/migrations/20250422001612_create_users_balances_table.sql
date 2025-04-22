-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_balances (
    user_id       UUID    NOT NULL PRIMARY KEY,
    balance       DECIMAL NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_balances;
-- +goose StatementEnd
