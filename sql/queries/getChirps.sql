-- name: GetAllChirps :many
SELECT *
FROM chirps
ORDER BY
  CASE WHEN $1::text = 'asc' THEN created_at END ASC,
  CASE WHEN $1::text  = 'desc' THEN created_at END DESC;