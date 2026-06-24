package sites

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/giorgiodots/dots-beacon/api/internal/server"
	"github.com/rs/zerolog"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-sites",
		Method:      http.MethodGet,
		Path:        "/sites",
		Summary:     "List sites",
		Description: "Returns all sites belonging to the authenticated user.",
		Tags:        []string{"sites"},
		Security:    []map[string][]string{{server.SecurityScheme: {}}},
	}, h.GetSites)

	huma.Register(api, huma.Operation{
		OperationID:   "create-site",
		Method:        http.MethodPost,
		Path:          "/sites",
		Summary:       "Create a site",
		Tags:          []string{"sites"},
		DefaultStatus: http.StatusCreated,
		Security:      []map[string][]string{{server.SecurityScheme: {}}},
	}, h.CreateSite)
}

// ListSitesOutput is the response body for GetSites. huma wraps the Body field
// as the JSON response and derives its schema from the struct tags.
type ListSitesOutput struct {
	Body struct {
		Sites []Site `json:"sites"`
	}
}

func (h *Handler) GetSites(ctx context.Context, _ *struct{}) (*ListSitesOutput, error) {
	logger := zerolog.Ctx(ctx)

	sites, err := h.svc.GetSites(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get sites")
		return nil, huma.Error500InternalServerError("internal error")
	}

	out := &ListSitesOutput{}
	out.Body.Sites = sites
	return out, nil
}

// CreateSiteInput carries the request body for CreateSite; huma validates it
// against the schema before the handler runs.
type CreateSiteInput struct {
	Body CreateSiteBody
}

type CreateSiteOutput struct {
	Body struct {
		Site Site `json:"site"`
	}
}

func (h *Handler) CreateSite(ctx context.Context, in *CreateSiteInput) (*CreateSiteOutput, error) {
	logger := zerolog.Ctx(ctx)

	site, err := h.svc.CreateSite(ctx, in.Body.Name)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create site")
		return nil, huma.Error422UnprocessableEntity("could not create site")
	}

	out := &CreateSiteOutput{}
	out.Body.Site = site
	return out, nil
}
