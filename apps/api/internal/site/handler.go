package site

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/giorgio-dots/dots-beacon-internal/auth"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
	"github.com/google/uuid"
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
	r.PATCH("/sites/:id", h.setIsOn)
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

type createSiteRequest struct {
	Name *string `json:"name" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Populated by auth.Middleware; falls back to "anonymous" when auth is off.
	actor, ok := auth.UserFromContext(c)
	if !ok {
		actor = auth.User{Subject: "anonymous", Username: "anonymous"}
	}
	created, err := h.svc.Create(c.Request.Context(), *req.Name, actor)
	if err != nil {
		telemetry.Log().Error().Ctx(c.Request.Context()).Err(err).Msg("failed to create site")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": created.ID})
}

// setIsOnRequest is the PATCH body. IsOn is a pointer so "field omitted" is
// distinguishable from "false" — binding:"required" rejects a missing field.
type setIsOnRequest struct {
	IsOn *bool `json:"isOn" binding:"required"`
}

func (h *Handler) setIsOn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site id"})
		return
	}

	var req setIsOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "isOn is required"})
		return
	}

	// Populated by auth.Middleware; falls back to "anonymous" when auth is off.
	actor, ok := auth.UserFromContext(c)
	if !ok {
		actor = auth.User{Subject: "anonymous", Username: "anonymous"}
	}

	updated, err := h.svc.SetOn(c.Request.Context(), id, *req.IsOn, actor)
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		return
	case err != nil:
		telemetry.Log().Error().Ctx(c.Request.Context()).Err(err).Msg("failed to set site state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"site": updated})
}
