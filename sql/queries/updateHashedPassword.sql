-- name: UpdatePassword :one
UPDATE users
SET updated_at = NOW(),
    hashed_password = $2
WHERE id = $1
RETURNING *;