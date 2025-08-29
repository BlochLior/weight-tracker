-- name: AddWeight :one
INSERT INTO weights (
    weight, date, unit, note
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;