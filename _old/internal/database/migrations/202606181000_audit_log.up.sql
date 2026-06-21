CREATE TABLE
    audit_log (
        id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        occurred_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        -- actor_id is TEXT (not UUID) so non-Keycloak / system actors fit too.
        actor_id TEXT NOT NULL,
        actor_name TEXT NOT NULL,
        action TEXT NOT NULL, -- e.g. 'site.toggled'
        target_type TEXT NOT NULL, -- e.g. 'site'
        target_id UUID NOT NULL,
        -- links the audit entry back to its trace (Tempo) + logs (Loki).
        trace_id TEXT,
        metadata JSONB NOT NULL DEFAULT '{}'::jsonb
    );

-- "what happened to this resource?" / "what did this user do?" / recent activity.
CREATE INDEX idx_audit_log_target ON audit_log (target_type, target_id, occurred_at DESC);

CREATE INDEX idx_audit_log_actor ON audit_log (actor_id, occurred_at DESC);

CREATE INDEX idx_audit_log_occurred_at ON audit_log (occurred_at DESC);
