package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/trace"
)

// Repository is the data-access layer for audit entries.
type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

// WithTx returns a Repository bound to the given (transaction-scoped) Queries,
// so a caller can record an audit entry inside the same transaction as the
// change it describes. See database.Tx.
func (r *Repository) WithTx(q *db.Queries) *Repository {
	return &Repository{q: q}
}

// Record persists one audit entry. The trace_id is taken from the active span in
// ctx (if any), so the entry links back to its trace and logs automatically.
func (r *Repository) Record(ctx context.Context, e Entry) error {
	metadata := e.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	targetID, err := uuid.Parse(e.TargetID)
	if err != nil {
		return fmt.Errorf("parse audit target id %q: %w", e.TargetID, err)
	}

	traceID := pgtype.Text{}
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		traceID = pgtype.Text{String: sc.TraceID().String(), Valid: true}
	}

	return r.q.CreateAuditEntry(ctx, db.CreateAuditEntryParams{
		ActorID:    e.ActorID,
		ActorName:  e.ActorName,
		Action:     e.Action,
		TargetType: e.TargetType,
		TargetID:   pgtype.UUID{Bytes: targetID, Valid: true},
		TraceID:    traceID,
		Metadata:   raw,
	})
}

// List returns the most recent audit entries, newest first.
func (r *Repository) List(ctx context.Context, limit int32) ([]Entry, error) {
	rows, err := r.q.ListAuditEntries(ctx, limit)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, toDomain(row))
	}
	return entries, nil
}

// toDomain maps a DB row to the domain model.
func toDomain(row db.AuditLog) Entry {
	var metadata map[string]any
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &metadata)
	}
	return Entry{
		ID:         row.ID,
		OccurredAt: row.OccurredAt.Time,
		ActorID:    row.ActorID,
		ActorName:  row.ActorName,
		Action:     row.Action,
		TargetType: row.TargetType,
		TargetID:   uuid.UUID(row.TargetID.Bytes).String(),
		TraceID:    row.TraceID.String,
		Metadata:   metadata,
	}
}
