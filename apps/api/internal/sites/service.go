package sites

import (
	"context"

	"github.com/giorgiodots/dots-beacon/package/database/db"
	"github.com/google/uuid"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) GetSites(ctx context.Context) ([]Site, error) {
	rows, err := s.q.GetSites(ctx)
	if err != nil {
		return nil, err
	}

	sites := make([]Site, 0, len(rows))
	for _, row := range rows {
		sites = append(sites, toDomain(row))
	}
	return sites, nil
}

func toDomain(site db.Site) Site {
	return Site{
		ID:   uuid.UUID(site.ID.Bytes),
		Name: site.Name,
		IsOn: site.IsOn,
	}
}
