-- name: CreateSite :one
INSERT INTO
    site (
        name,
        is_on
    )
VALUES
    ($1, false)
RETURNING
    id, name, is_on;
