-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    token,
    created_at,
    updated_at,
    expires_at,
    user_id
) VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;


-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = CURRENT_TIMESTAMP,
    revoked_at = $2
WHERE token = $1;