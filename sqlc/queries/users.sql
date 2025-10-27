-- name: CreateUser :one
INSERT INTO users (full_name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1
    AND deleted_at IS NULL;
-- name: UpdateUser :one
UPDATE users
SET full_name = $2,
    email = $3,
    updated_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1;
-- name: ListUsers :many
SELECT *
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;