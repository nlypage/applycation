-- name: CreateOwnerSession :one
INSERT INTO owner_sessions (
  owner_id,
  session_token_hash,
  user_agent,
  ip_address,
  expires_at
) VALUES (
  sqlc.arg(owner_id),
  sqlc.arg(session_token_hash),
  sqlc.arg(user_agent),
  sqlc.arg(ip_address),
  sqlc.arg(expires_at)
)
RETURNING *;

-- name: GetOwnerSessionByTokenHash :one
SELECT *
FROM owner_sessions
WHERE session_token_hash = sqlc.arg(session_token_hash)
LIMIT 1;

-- name: TouchOwnerSession :one
UPDATE owner_sessions
SET
  last_seen_at = now()
WHERE session_token_hash = sqlc.arg(session_token_hash)
RETURNING *;

-- name: RevokeOwnerSession :exec
UPDATE owner_sessions
SET
  revoked_at = now()
WHERE session_token_hash = sqlc.arg(session_token_hash);

-- name: DeleteExpiredOwnerSessions :execrows
DELETE FROM owner_sessions
WHERE expires_at < sqlc.arg(now_ts)
   OR revoked_at IS NOT NULL;
