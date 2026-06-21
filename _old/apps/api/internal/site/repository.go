package site

import (
	"context"
	"errors"

	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

// WithTx returns a Repository bound to the given (transaction-scoped) Queries.
// The service uses it to run a write together with its audit entry atomically.
func (r *Repository) WithTx(q *db.Queries) *Repository {
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

func (r *Repository) Create(ctx context.Context, name string) (Site, error) {
	row, err := r.q.CreateSite(ctx, name)
	if err != nil {
		return Site{}, err
	}
	return toDomain(row), nil
}

// SetIsOn turns a site on or off and returns the updated site.
func (r *Repository) SetIsOn(ctx context.Context, id uuid.UUID, isOn bool) (Site, error) {
	row, err := r.q.SetSiteIsOn(ctx, db.SetSiteIsOnParams{
		ID:   pgtype.UUID{Bytes: id, Valid: true},
		IsOn: isOn,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Site{}, ErrNotFound
		}
		return Site{}, err
	}
	return toDomain(row), nil
}

// toDomain maps a DB row to the domain model.
func toDomain(row db.Site) Site {
	return Site{
		ID:   uuid.UUID(row.ID.Bytes), // pgtype.UUID.Bytes is a [16]byte
		Name: row.Name,
		IsOn: row.IsOn,
	}
}
