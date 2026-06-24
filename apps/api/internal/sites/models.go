package sites

import "github.com/google/uuid"

type Site struct {
	ID   uuid.UUID `json:"ID" doc:"Unique site identifier"`
	Name string    `json:"name" doc:"Display name of the site"`
	IsOn bool      `json:"is_on" doc:"Whether the site is currently enabled"`
}

type CreateSiteBody struct {
	Name string `json:"name" minLength:"4" doc:"Display name of the site"`
}
