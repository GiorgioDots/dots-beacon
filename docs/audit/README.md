# Auditing

dots-beacon keeps **auditing** and **observability** separate on purpose — they
answer different questions and have different correctness requirements.

| | Observability (traces/logs/metrics) | Audit log |
|---|---|---|
| Question | "Is the system healthy? why is this slow/failing?" | "Who did what to which resource, when?" |
| Store | Tempo / Loki / Mimir (S3) | **Postgres** (`audit_log`) |
| Authoritative? | No — sampled, best-effort, may drop | **Yes** — must not be lost |
| Retention | Days/weeks (cost-driven) | Months/years (policy-driven) |

The rule of thumb: **logs are allowed to lie** (sampled, batched, expired). An
audit entry that says "user X turned site Y on" must be exact and durable, so it
is written to Postgres **in the same transaction as the change itself**.

## How it works

- **Table:** [`audit_log`](../../internal/database/migrations/202606181000_audit_log.up.sql)
  — `actor_id`, `actor_name`, `action`, `target_type`, `target_id`, `trace_id`,
  `metadata` (jsonb), `occurred_at`.
- **The bridge to observability:** every row stores the **`trace_id`** of the
  request that caused it (taken from the active span). So from an audit entry you
  can jump straight to the full trace in Tempo and the correlated logs in Loki —
  business record and deep technical detail, without conflating the two stores.
- **Atomicity:** the change and its audit entry commit together via
  [`database.Tx.Run`](../../internal/database/tx.go). If the audit write fails,
  the change rolls back too — the log can never miss something that happened.

```
PATCH /sites/:id ──► site.Handler ──► site.Service.SetOn
                                         │  database.Tx.Run(ctx, func(q) {
                                         │     site.Repo.WithTx(q).SetIsOn(...)   ─┐ one
                                         │     audit.Repo.WithTx(q).Record(...)   ─┘ transaction
                                         │  })
                                         ▼
                                    Postgres: site UPDATE + audit_log INSERT (atomic)
```

## What to audit (and what not to)

- **Audit:** state-changing business actions — create/update/delete/toggle,
  permission changes, logins you care about. The test: *would you be unhappy if
  this record went missing?* If yes, it's an audit event.
- **Don't audit:** reads, request timing, errors, DB latency — that's all already
  covered by observability. Don't duplicate it into `audit_log`.

## Recording an entry from a feature

Do it in the **service layer**, inside the transaction that performs the change
(not in middleware — middleware sees HTTP, not domain intent like "toggled on").
See [`site.Service.SetOn`](../../apps/api/internal/site/service.go):

```go
err := s.tx.Run(ctx, func(q *db.Queries) error {
    updated, err = s.repo.WithTx(q).SetIsOn(ctx, id, isOn)
    if err != nil {
        return err
    }
    return s.audit.WithTx(q).Record(ctx, audit.Entry{
        ActorID:    actor.Subject,   // from auth.UserFromContext
        ActorName:  actor.Username,
        Action:     "site.toggled",
        TargetType: "site",
        TargetID:   id.String(),
        Metadata:   map[string]any{"isOn": isOn},
    })
})
```

`Record` fills `trace_id` from the context automatically. The actor comes from the
authenticated `auth.User`; when auth is disabled (dev) it falls back to
`anonymous`.

## Reading the log

`GET /audit-log?limit=N` — **admin only** (checks `User.HasRole("admin")`). This
is the one place authorization is enforced; it builds on the roles from
[`internal/auth`](../../internal/auth/), which only authenticates.

```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" http://localhost:8080/audit-log
```

## Notes / future

- `metadata` is free-form jsonb. It currently stores the new value
  (`{"isOn": true}`); enrich it with before/after (`{"isOn": {"from": …, "to": …}}`)
  when a richer diff is useful.
- The `audit` package lives with the API ([`apps/api/internal/audit`](../../apps/api/internal/audit/)).
  If other services need to write entries later, promote it to the shared
  `internal/` module — same code, different location.
- `audit_log` is append-only by convention; never `UPDATE`/`DELETE` it in app code.
