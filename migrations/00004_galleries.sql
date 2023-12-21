-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS galleries
(
    id      SERIAL PRIMARY KEY,
    user_id INT REFERENCES users (id),
    title   TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS galleries;
-- +goose StatementEnd
