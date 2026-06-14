package site

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
)

// Handler exposes the site endpoints over HTTP. It depends only on the service.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the site routes on the given router/group. The server
// passes the group these routes should live under (e.g. the authenticated one).
func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/sites", h.list)
}

func (h *Handler) list(c *gin.Context) {
	sites, err := h.svc.List(c.Request.Context())
	if err != nil {
		telemetry.Log().Error().Ctx(c.Request.Context()).Err(err).Msg("failed to list sites")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sites": sites})
}
