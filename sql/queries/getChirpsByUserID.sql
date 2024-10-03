-- name: GetChirpsByUserID :many
SELECT *
FROM chirps
WHERE user_id = $1
ORDER BY
  CASE WHEN $2::text = 'asc' THEN created_at END ASC,
  CASE WHEN $2::text  = 'desc' THEN created_at END DESC;