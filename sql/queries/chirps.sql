-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),  -- Generates a new UUID
    NOW(),              -- Sets created_at to the current timestamp
    NOW(),              -- Sets updated_at to the current timestamp
    $1,                 -- The body, passed in by the application
    $2                  -- The user_id, passed in by the application
)
RETURNING *;