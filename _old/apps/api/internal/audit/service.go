package audit

import "context"

const (
	defaultListLimit = 100
	maxListLimit     = 500
)

// Service holds business logic for reading the audit log. Writing is done by
// other features inside their own transactions via Repository.Record, so there's
// no write method here.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns recent audit entries, clamping the limit to a sane range.
func (s *Service) List(ctx context.Context, limit int32) ([]Entry, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	return s.repo.List(ctx, limit)
}
