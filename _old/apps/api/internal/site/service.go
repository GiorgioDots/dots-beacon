package site

import (
	"context"

	"github.com/giorgio-dots/dots-beacon-api/internal/audit"
	"github.com/giorgio-dots/dots-beacon-internal/auth"
	"github.com/giorgio-dots/dots-beacon-internal/database"
	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/google/uuid"
)

// Service holds the business logic for sites. Handlers call the service; the
// service calls the repository. Validation, authorization rules, and
// orchestration across repositories belong here — keep handlers thin and the
// repository dumb.
//
// It also depends on the audit recorder and the transaction runner so a write
// and its audit entry commit atomically (see SetOn).
type Service struct {
	repo  *Repository
	audit *audit.Repository
	tx    *database.Tx
}

func NewService(repo *Repository, auditRepo *audit.Repository, tx *database.Tx) *Service {
	return &Service{repo: repo, audit: auditRepo, tx: tx}
}

// List returns all sites. (No business rules yet — this is where they'd go.)
func (s *Service) List(ctx context.Context) ([]Site, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, name string, actor auth.User) (Site, error) {
	var created Site
	err := s.tx.Run(ctx, func(q *db.Queries) error {
		var err error
		created, err = s.repo.WithTx(q).Create(ctx, name)
		if err != nil {
			return err
		}

		return s.audit.WithTx(q).Record(ctx, audit.Entry{
			ActorID:    actor.Subject,
			ActorName:  actor.Username,
			Action:     "site.toggled",
			TargetType: "site",
			TargetID:   created.ID.String(),
			Metadata:   nil,
		})
	})
	return created, err
}

// SetOn turns a site on or off and records an audit entry for the change. Both
// writes run in one transaction: if the audit write fails, the toggle rolls back
// too, so the audit log can never miss a change that actually happened.
func (s *Service) SetOn(ctx context.Context, id uuid.UUID, isOn bool, actor auth.User) (Site, error) {
	var updated Site
	err := s.tx.Run(ctx, func(q *db.Queries) error {
		var err error
		updated, err = s.repo.WithTx(q).SetIsOn(ctx, id, isOn)
		if err != nil {
			return err
		}

		return s.audit.WithTx(q).Record(ctx, audit.Entry{
			ActorID:    actor.Subject,
			ActorName:  actor.Username,
			Action:     "site.toggled",
			TargetType: "site",
			TargetID:   id.String(),
			Metadata:   map[string]any{"isOn": isOn},
		})
	})
	return updated, err
}
