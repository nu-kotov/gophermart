-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    number        BIGINT                   NOT NULL PRIMARY KEY,
    user_id       UUID                     NOT NULL,
    status        TEXT                     NOT NULL,
    accrual       DECIMAL(12, 2)               NULL,
    uploaded_at   TIMESTAMP WITH TIME ZONE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
