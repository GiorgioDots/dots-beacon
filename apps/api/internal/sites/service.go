package sites

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/giorgiodots/dots-beacon/api/internal/models"
	"github.com/giorgiodots/dots-beacon/package/database/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) getSites(c *gin.Context) {
	rows, err := s.q.GetSites(c)
	logger := zerolog.Ctx(c.Request.Context())
	if err != nil {
		logger.Error().Err(err).Msg("failed to get sites")
		c.JSON(http.StatusInternalServerError, models.NewResponse("Internal error", nil))
		return
	}

	sites := make([]Site, 0, len(rows))
	for _, row := range rows {
		sites = append(sites, toDomain(row))
	}
	c.JSON(http.StatusOK, models.NewResponse("ok", gin.H{"sites": sites}))
}

func toDomain(site db.Site) Site {
	return Site{
		ID:   uuid.UUID(site.ID.Bytes),
		Name: site.Name,
		IsOn: site.IsOn,
	}
}
