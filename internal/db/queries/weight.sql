-- name: AddWeightEntry :one
INSERT INTO weights (
    id, date, weight, weight_unit, note
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: AddWeightsUsersEntry :exec
INSERT INTO users_weights (
    user_id,
    weight_id
) VALUES (
    ?, ?
); 

-- name: DeleteWeightData :exec
DELETE FROM weights;

-- name: GetAllUserWeights :many
SELECT
    weights.id, 
    weights.date, 
    weights.weight, 
    weights.weight_unit, 
    weights.note
FROM 
    users
JOIN
    users_weights ON users.id = users_weights.user_id
JOIN
    weights ON users_weights.weight_id = weights.id
WHERE
    users.username = ?;

-- name: GetWeightEntry :one
SELECT
    id, date, weight, weight_unit, note
FROM 
    weights
WHERE
    id = ?;

-- name: DeleteUserWeight :exec
DELETE FROM 
    weights
WHERE
    id = ?;

-- name: DeleteAllUserWeights :exec
DELETE FROM 
    weights
WHERE
    id IN (
        SELECT users_weights.weight_id
        FROM users_weights
        WHERE users_weights.user_id = ?
    );

-- name: GetUserIDFromWeightID :one
SELECT
    user_id
FROM
    users_weights
WHERE
    weight_id = ?;
