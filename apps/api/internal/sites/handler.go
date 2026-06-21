package sites

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	grp := r.Group("/sites")
	grp.GET("/", h.svc.getSites)
}
