package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// Scopes defines the admin API's capability boundary over the scopes store.
type Scopes interface {
	Define(ctx context.Context, applicationID, name, description string) (*models.Scope, error)
	GetByID(ctx context.Context, id string) (*models.Scope, error)
	ListByApplication(ctx context.Context, applicationID string) ([]*models.Scope, error)
	// ListAll retrieves every scope across all applications.
	ListAll(ctx context.Context) ([]*models.Scope, error)
	// Search runs the cross-application search over scopes.
	Search(ctx context.Context, params stores.ScopeSearchParams) (*stores.ScopeSearchResult, error)
	Delete(ctx context.Context, id string) error
}
