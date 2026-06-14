package site

import (
	"context"

	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/google/uuid"
)

// Repository is the data-access layer for sites. It wraps the sqlc-generated
// queries and maps DB rows to domain models, so nothing above it depends on
// pgx/sqlc types.
type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

// List returns all sites.
func (r *Repository) List(ctx context.Context) ([]Site, error) {
	rows, err := r.q.ListSites(ctx)
	if err != nil {
		return nil, err
	}

	sites := make([]Site, 0, len(rows))
	for _, row := range rows {
		sites = append(sites, toDomain(row))
	}
	return sites, nil
}

// toDomain maps a DB row to the domain model.
func toDomain(row db.Site) Site {
	return Site{
		ID:   uuid.UUID(row.ID.Bytes), // pgtype.UUID.Bytes is a [16]byte
		Name: row.Name,
		IsOn: row.IsOn,
	}
}
