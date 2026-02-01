-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByRefreshTokenHash :one
SELECT * FROM sessions
WHERE refresh_token_hash = $1 AND expires_at > NOW();

-- name: UpdateSessionLastUsed :exec
UPDATE sessions
SET last_used_at = NOW()
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE refresh_token_hash = $1;

-- name: DeleteAllUserSessions :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= NOW();

-- name: BlacklistToken :exec
INSERT INTO token_blacklist (token_jti, user_id, expires_at, reason)
VALUES ($1, $2, $3, $4);

-- name: IsTokenBlacklisted :one
SELECT EXISTS(
    SELECT 1 FROM token_blacklist
    WHERE token_jti = $1 AND expires_at > NOW()
) AS is_blacklisted;

-- name: DeleteExpiredBlacklistedTokens :exec
DELETE FROM token_blacklist
WHERE expires_at <= NOW();

-- name: GetActiveSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = $1 AND expires_at > NOW()
ORDER BY created_at DESC;
