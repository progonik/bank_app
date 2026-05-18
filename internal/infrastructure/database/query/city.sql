-- name: CreateCity :one
INSERT INTO cities (name)
VALUES ($1)
RETURNING id, name, created_at;

-- name: GetCityByID :one
SELECT id, name, created_at
FROM cities
WHERE id = $1;

-- name: UpdateCity :one
UPDATE cities
SET name = $2
WHERE id = $1
RETURNING id, name, created_at;

-- name: DeleteCity :exec
DELETE FROM cities
WHERE id = $1;
