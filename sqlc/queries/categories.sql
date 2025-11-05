-- name: CreateCategory :one
INSERT INTO categories (user_id, name, description, color, icon)
VALUES($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetCategoryByID :one
SELECT *
FROM categories
WHERE id = $1
    AND deleted_at IS NULL;
-- name: GetCategoriesByUserID :many
SELECT *
FROM categories
WHERE user_id = $1
    AND deleted_at IS NULL
ORDER BY created_at DESC;
-- name: UpdateCategory :one
UPDATE categories
SET name = COALESCE($2, description),
    description = COALESCE($3, description),
    color = COALESCE($4, color),
    icon = COALESCE($5, icon),
    is_active = COALESCE($6, is_active),
    updated_at = NOW()
WHERE id = $1
    AND deleted_at IS NULL
RETURNING *;
-- name: DeletedCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE id = $1;
-- name: GetCategoryByUserIDAndName :one
SELECT *
FROM categories
WHERE user_id = $1
    AND name = $2
    AND deleted_at is NULL;