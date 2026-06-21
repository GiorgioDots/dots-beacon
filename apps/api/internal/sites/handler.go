package sites

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/giorgiodots/dots-beacon/api/internal/respond"
	"github.com/rs/zerolog"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/sites")
	grp.GET("/", h.GetSites)
}

func (h *Handler) GetSites(c *gin.Context) {
	logger := zerolog.Ctx(c.Request.Context())
	sites, err := h.svc.GetSites(c.Request.Context())
	if err != nil {
		logger.Error().Err(err).Msg("failed to get sites")
		respond.Err(c, http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"sites": sites})
}
