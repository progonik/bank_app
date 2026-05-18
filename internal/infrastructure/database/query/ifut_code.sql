-- name: CreateIfutCode :one
INSERT INTO ifut_codes (name)
VALUES ($1)
RETURNING id, name, created_at;

-- name: GetIfutCodeByID :one
SELECT id, name, created_at
FROM ifut_codes
WHERE id = $1;

-- name: GetIfutCodeByName :one
SELECT id, name, created_at
FROM ifut_codes
WHERE name = $1;
