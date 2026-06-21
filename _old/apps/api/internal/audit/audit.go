// Package audit records authoritative "who did what, when" entries for
// state-changing business actions. Unlike telemetry (traces/logs/metrics, which
// are sampled, best-effort, and short-lived), audit entries are durable and must
// not be lost — so they live in Postgres and are written in the same transaction
// as the change they describe. Each entry carries the trace_id, so an audit row
// links straight back to the full trace (Tempo) and correlated logs (Loki).
package audit

import "time"

// Entry is one audit record. Callers fill the actor/action/target/metadata; the
// repository stamps occurred_at (DB default) and the trace_id (from context).
type Entry struct {
	ID         int64          `json:"id"`
	OccurredAt time.Time      `json:"occurredAt"`
	ActorID    string         `json:"actorId"`    // auth User.Subject (or "anonymous"/system)
	ActorName  string         `json:"actorName"`  // auth User.Username, snapshotted
	Action     string         `json:"action"`     // e.g. "site.toggled"
	TargetType string         `json:"targetType"` // e.g. "site"
	TargetID   string         `json:"targetId"`   // the affected resource's id
	TraceID    string         `json:"traceId,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"` // e.g. {"isOn": {"from": false, "to": true}}
}
