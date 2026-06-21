-- name: CreateAuditEntry :exec
INSERT INTO
    audit_log (
        actor_id,
        actor_name,
        action,
        target_type,
        target_id,
        trace_id,
        metadata
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7);
