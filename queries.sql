-- name: AddWeight :one
INSERT INTO weights (
    weight, date, unit, note
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;

-- name: ListWeightsDateAsc :many
SELECT * FROM weights
WHERE
    (@start_date IS NULL OR date >= @start_date)
    AND (@end_date IS NULL OR date <= @end_date)
ORDER BY date ASC
LIMIT @row_limit;

-- name: ListWeightsDateDesc :many
SELECT * FROM weights
WHERE
    (@start_date IS NULL OR date >= @start_date)
    AND (@end_date IS NULL OR date <= @end_date)
ORDER BY date DESC
LIMIT @row_limit;

-- name: ListWeightsWeightAsc :many
SELECT * FROM weights
WHERE
    (@start_date IS NULL OR date >= @start_date)
    AND (@end_date IS NULL OR date <= @end_date)
ORDER BY weight ASC
LIMIT @row_limit;

-- name: ListWeightsWeightDesc :many
SELECT * FROM weights
WHERE
    (@start_date IS NULL OR date >= @start_date)
    AND (@end_date IS NULL OR date <= @end_date)
ORDER BY weight DESC
LIMIT @row_limit;

-- name: GetWeight :one
SELECT * FROM weights WHERE id = ?;

-- name: DeleteWeight :exec
DELETE FROM weights WHERE id = ?;

-- name: UpdateWeight :one
UPDATE weights SET 
    weight = ?,
    date = ?,
    unit = ?,
    note = ?,
    user_id = ?
WHERE id = ?
RETURNING *;