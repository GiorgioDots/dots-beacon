package sites

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/giorgiodots/dots-beacon/api/internal/models"
	"github.com/giorgiodots/dots-beacon/package/database/db"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/sites")
	grp.GET("/", h.getSites)
}

func (h *Handler) getSites(c *gin.Context) {
	rows, err := h.svc.q.GetSites(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewResponse("Internal error", nil))
		return
	}

	sites := make([]Site, 0, len(rows))
	for _, row := range rows {
		sites = append(sites, toDomain(row))
	}
}

func toDomain(site db.Site) Site {
	return Site{
		ID:   uuid.UUID(site.ID.Bytes),
		Name: site.Name,
		IsOn: site.IsOn,
	}
}
