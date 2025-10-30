-- name: CreateTransaction :one
INSERT INTO transactions (
        user_id,
        account_id,
        category_id,
        type,
        amount,
        description,
        status,
        date
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        COALESCE($7, 'completed'),
        $8
    )
RETURNING *;
-- name: GetTransactionByID :one
SELECT *
FROM transactions
WHERE id = $1
    AND deleted_at IS NULL
LIMIT 1;
-- name: ListTransactionsByAccount :many
SELECT *
FROM transactions
WHERE account_id = $1
    AND deleted_at IS NULL
ORDER BY date DESC,
    created_at DESC;
-- name: ListTransactionsByUser :many
SELECT *
FROM transactions
WHERE user_id = $1
    AND deleted_at IS NULL
ORDER BY date DESC,
    created_at DESC;
-- name: ListTransactionsByPeriod :many
SELECT *
FROM transactions
WHERE user_id = $1
    AND date >= $2
    AND date < $3
    AND deleted_at IS NULL
ORDER BY date DESC,
    created_at DESC;
-- name: UpdateTransaction :one
UPDATE transactions
SET account_id = COALESCE(sqlc.narg('account_id'), account_id),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    type = COALESCE(sqlc.narg('type'), type),
    amount = COALESCE(sqlc.narg('amount'), amount),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    date = COALESCE(sqlc.narg('date'), date),
    updated_at = NOW()
WHERE id = @id
    AND deleted_at IS NULL
RETURNING *;
-- name: SoftDeleteTransaction :exec
UPDATE transactions
SET deleted_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL;