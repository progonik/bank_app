-- name: CreateINN :one
INSERT INTO inns (name)
VALUES ($1)
RETURNING id, name, created_at;

-- name: GetINNByID :one
SELECT id, name, created_at
FROM inns
WHERE id = $1;

-- name: GetINNByName :one
SELECT id, name, created_at
FROM inns
WHERE name = $1;
