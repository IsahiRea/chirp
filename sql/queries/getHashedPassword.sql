-- name: GetHashPassByEmail :one
SELECT *
FROM users
WHERE email=$1;