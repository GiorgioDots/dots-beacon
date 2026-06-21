-- name: SetSiteIsOn :one
UPDATE site
SET
    is_on = $2
WHERE
    id = $1
RETURNING
    id, name, is_on;
