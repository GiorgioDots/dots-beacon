package audit

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/giorgio-dots/dots-beacon-internal/auth"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
)

// Handler exposes the audit-log read endpoint. It depends only on the service.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.GET("/audit-log", h.list)
}

// list returns recent audit entries. Reading the audit log is admin-only — this
// is the one bit of authorization in the example; it builds on auth's User.Roles.
func (h *Handler) list(c *gin.Context) {
	user, ok := auth.UserFromContext(c)
	if !ok || !user.HasRole("admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		return
	}

	var limit int32
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = int32(n)
		}
	}

	entries, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		telemetry.Log().Error().Ctx(c.Request.Context()).Err(err).Msg("failed to list audit entries")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}
