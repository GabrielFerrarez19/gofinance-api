-- name: CreateAccount :one
INSERT INTO accounts (
        user_id,
        name,
        type,
        balance,
        currency,
        description,
        is_active
    )
VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, TRUE))
RETURNING *;
-- name: GetAccountByID :one
SELECT *
FROM accounts
WHERE id = $1
    AND deleted_at IS NULL
LIMIT 1;
-- name: ListAccountsByUser :many
SELECT *
FROM accounts
WHERE user_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC;
-- name: UpdateAccount :one
UPDATE accounts
SET name = COALESCE($2, name),
    type = COALESCE($3, type),
    balance = COALESCE($4, balance),
    currency = COALESCE($5, currency),
    description = COALESCE($6, description),
    is_active = COALESCE($7, is_active),
    updated_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: SoftDeleteAccount :exec
UPDATE accounts
SET deleted_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL;
-- name: UpdateAccountBalance :one
UPDATE accounts
SET balance = $2,
    updated_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;