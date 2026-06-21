package sites

import "github.com/google/uuid"

type Site struct {
	ID   uuid.UUID `json:"ID"`
	Name string    `json:"name"`
	IsOn bool      `json:"is_on"`
}

type CreateSiteBody struct {
	Name string `json:"name" binding:"required,gt=3"`
}
