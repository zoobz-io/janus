package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Applications defines the contract for application operations on the public API surface.
type Applications interface {
	// GetApplication retrieves an application by ID.
	GetApplication(ctx context.Context, id string) (*models.Application, error)
	// ListApplications retrieves all applications.
	ListApplications(ctx context.Context) ([]*models.Application, error)
}
