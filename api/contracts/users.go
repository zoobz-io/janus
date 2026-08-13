// Package contracts defines interface boundaries for the public API surface.
package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Users defines the contract for user operations on the public API surface.
// All methods are scoped to the authenticated user — no cross-tenant access.
type Users interface {
	// GetUser retrieves a user by internal ID.
	GetUser(ctx context.Context, id string) (*models.User, error)
	// UpdateDisplayName updates the authenticated user's display name.
	UpdateDisplayName(ctx context.Context, id, displayName string) (*models.User, error)
}
