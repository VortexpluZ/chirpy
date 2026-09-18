-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: TruncateUsers :exec
TRUNCATE TABLE users CASCADE;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET updated_at = CURRENT_TIMESTAMP,
    email = $3,
    hashed_password = $2
WHERE id = $1
RETURNING *;


-- name: UpdateUserToRed :one
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;