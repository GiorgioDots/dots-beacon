-- name: ListAuditEntries :many
SELECT
    id,
    occurred_at,
    actor_id,
    actor_name,
    action,
    target_type,
    target_id,
    trace_id,
    metadata
FROM
    audit_log
ORDER BY
    occurred_at DESC,
    id DESC
LIMIT
    $1;
