package site

import "context"

// Service holds the business logic for sites. Handlers call the service; the
// service calls the repository. Validation, authorization rules, and
// orchestration across repositories belong here — keep handlers thin and the
// repository dumb.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all sites. (No business rules yet — this is where they'd go.)
func (s *Service) List(ctx context.Context) ([]Site, error) {
	return s.repo.List(ctx)
}
