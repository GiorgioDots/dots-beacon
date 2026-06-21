-- name: CreateSite :one
INSERT INTO sites (
    name, is_on
) values (
    $1, false
)
RETURNING *;