-- name: CreateSession :one
INSERT INTO user_active_sessions (user_id, device_id, refresh_token_hash, last_login_at)
VALUES ($1, $2, $3, NOW())
RETURNING id, user_id, device_id, refresh_token_hash, last_login_at, created_at;

-- name: GetSessionByRefreshTokenHash :one
SELECT id, user_id, device_id, refresh_token_hash, last_login_at, created_at
FROM user_active_sessions
WHERE refresh_token_hash = $1;

-- name: GetSessionByUserAndDevice :one
SELECT id, user_id, device_id, refresh_token_hash, last_login_at, created_at
FROM user_active_sessions
WHERE user_id = $1 AND device_id = $2;

-- name: DeleteSession :exec
DELETE FROM user_active_sessions
WHERE id = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM user_active_sessions
WHERE user_id = $1;

-- name: UpdateSessionToken :one
UPDATE user_active_sessions
SET refresh_token_hash = $2, last_login_at = NOW()
WHERE id = $1
RETURNING id, user_id, device_id, refresh_token_hash, last_login_at, created_at;
