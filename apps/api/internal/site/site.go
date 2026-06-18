// Package site is one feature slice: the domain model plus its repository,
// service, and HTTP handler. Copy this package's shape for new features
// (devices, users, ...). The flow is always: handler -> service -> repository.
package site

import (
	"errors"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a site does not exist. The handler maps it to 404.
var ErrNotFound = errors.New("site not found")

// Site is the domain model the API exposes. It is intentionally separate from
// the sqlc-generated DB row (db.Site) so the HTTP layer never depends on the
// database schema — the repository maps between them.
type Site struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	IsOn bool      `json:"isOn"`
}
