-- name: ListSites :many
SELECT
    *
FROM
    site
ORDER BY
    name;