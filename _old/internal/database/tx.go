package database

import (
	"context"
	"fmt"

	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tx runs work inside a single database transaction. Build one from the pool and
// share it with services that must commit several writes atomically — e.g. a
// domain change together with its audit entry. It's the unit-of-work seam:
// repositories stay unaware of transactions; the service decides the boundary.
type Tx struct {
	pool *pgxpool.Pool
}

func NewTx(pool *pgxpool.Pool) *Tx {
	return &Tx{pool: pool}
}

// Run executes fn within one transaction. fn receives tx-scoped Queries; rebind
// any repository onto them with its WithTx method so every write in fn hits the
// same transaction. The transaction commits when fn returns nil and rolls back
// on any error (or panic), so either all writes land or none do.
func (t *Tx) Run(ctx context.Context, fn func(q *db.Queries) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op once committed; safety net on early return/panic

	if err := fn(db.New(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
