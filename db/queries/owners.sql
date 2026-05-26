-- name: CreateOwner :one
INSERT INTO owners (
  password_hash
) VALUES (
  sqlc.arg(password_hash)
)
RETURNING *;

-- name: GetOwnerByID :one
SELECT *
FROM owners
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetSingleOwner :one
SELECT *
FROM owners
ORDER BY created_at ASC
LIMIT 1;

-- name: UpdateOwnerPassword :one
UPDATE owners
SET
  password_hash = sqlc.arg(password_hash),
  password_changed_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;
