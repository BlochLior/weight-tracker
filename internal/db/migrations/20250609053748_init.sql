-- +goose Up
-- +goose StatementBegin
CREATE TABLE weights (
    id TEXT PRIMARY KEY NOT NULL,
    date TEXT NOT NULL,
    weight REAL NOT NULL,
    weight_unit TEXT NOT NULL,
    note TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE weights;
-- +goose StatementEnd
