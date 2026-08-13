package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Authorizations defines the contract for resolving a user's authorization
// for an application: the tenants that entitle them to it, with membership
// role, application roles, and effective scopes.
type Authorizations interface {
	// ForApplication resolves the application by slug and the caller's
	// authorized tenants for it. The tenant list is empty when the user has
	// no entitlements; an unknown slug is an error.
	ForApplication(ctx context.Context, userID, appSlug string) (*models.Application, []models.AuthorizedTenant, error)
}
