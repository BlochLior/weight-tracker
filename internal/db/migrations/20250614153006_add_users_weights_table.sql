-- +goose Up
-- +goose StatementBegin
CREATE TABLE users_weights (
    user_id TEXT NOT NULL,
    weight_id TEXT NOT NULL,
    PRIMARY KEY (user_id, weight_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (weight_id) REFERENCES weights (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users_weights;
-- +goose StatementEnd
