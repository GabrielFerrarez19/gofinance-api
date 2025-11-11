-- name: CreateReport :one
INSERT INTO reports (user_id, type, title, description, start_date, end_date, data)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetReportByID :one
SELECT *
FROM reports
WHERE id = $1;

-- name: ListReportsByUser :many
SELECT *
FROM reports
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateReport :one
UPDATE reports
SET title = COALESCE($2, title),
    description = COALESCE($3, description),
    data = COALESCE($4, data),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteReport :exec
DELETE FROM reports
WHERE id = $1;