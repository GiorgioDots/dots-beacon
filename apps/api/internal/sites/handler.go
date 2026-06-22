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

func (h *Handler) RegisterRoutes(r gin.IRouter, authMW gin.HandlerFunc) {
	grp := r.Group("/sites")
	authenticated := grp.Use(authMW)
	authenticated.GET("/", h.GetSites)
	authenticated.POST("/", h.CreateSite)
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

func (h *Handler) CreateSite(c *gin.Context) {
	logger := zerolog.Ctx(c.Request.Context())
	var body CreateSiteBody
	if err := c.ShouldBind(&body); err != nil {
		logger.Error().Err(err).Msg("create site body not valid")
		respond.Err(c, http.StatusUnprocessableEntity, "invalid request")
		return
	}

	site, err := h.svc.CreateSite(c.Request.Context(), body.Name)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create site")
		respond.Err(c, http.StatusUnprocessableEntity, "invalid error")
	}

	c.JSON(http.StatusCreated, gin.H{"craeted": site})
}
