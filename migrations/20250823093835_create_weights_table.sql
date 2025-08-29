-- +goose Up
CREATE TABLE weights (
    id INTEGER PRIMARY KEY NOT NULL,
    weight REAL NOT NULL,
    date TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unit TEXT NOT NULL DEFAULT 'kg',
    note TEXT,
    user_id INTEGER
);

-- +goose Down
DROP TABLE weights;