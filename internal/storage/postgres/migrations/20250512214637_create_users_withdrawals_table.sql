-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals (
    number       BIGINT                   NOT NULL PRIMARY KEY,
    user_id      UUID                     NOT NULL,
    sum          DECIMAL(12, 2)               NULL,
    withdrawn_at TIMESTAMP WITH TIME ZONE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
