-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),  -- Generates a new UUID
    NOW(),              -- Sets created_at to the current timestamp
    NOW(),              -- Sets updated_at to the current timestamp
    $1                  -- The email, passed in by the application
)
RETURNING *;
