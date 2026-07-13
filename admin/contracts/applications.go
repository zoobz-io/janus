// Package contracts defines the interface boundaries for the admin API surface.
//
// Stores are shared with the public API, but each API sees only the methods
// its own contract exposes. The admin contracts are broader than the public
// ones — e.g. admin can create and update applications, the public API cannot.
package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Applications defines the admin API's capability boundary over the applications store.
type Applications interface {
	// GetApplication retrieves an application by ID.
	GetApplication(ctx context.Context, id string) (*models.Application, error)
	// ListAll retrieves every application regardless of status.
	ListAll(ctx context.Context) ([]*models.Application, error)
	// CreateApplication registers a new application.
	CreateApplication(ctx context.Context, name, slug string) (*models.Application, error)
	// Update updates an application's name and status.
	Update(ctx context.Context, id, name string, status models.ApplicationStatus) (*models.Application, error)
}
