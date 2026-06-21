package sites

import "github.com/giorgiodots/dots-beacon/package/database/db"

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}
