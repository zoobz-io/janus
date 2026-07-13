package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Scopes defines the admin API's capability boundary over the scopes store.
type Scopes interface {
	Define(ctx context.Context, applicationID, name, description string) (*models.Scope, error)
	GetByID(ctx context.Context, id string) (*models.Scope, error)
	ListByApplication(ctx context.Context, applicationID string) ([]*models.Scope, error)
	Delete(ctx context.Context, id string) error
}
