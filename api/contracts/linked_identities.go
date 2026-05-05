package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// LinkedIdentities defines the contract for linked identity operations on the public API surface.
type LinkedIdentities interface {
	// ListByUser retrieves all linked identities for a user.
	ListByUser(ctx context.Context, userID string) ([]*models.LinkedIdentity, error)
	// Unlink removes a linked identity by ID, scoped to the user.
	Unlink(ctx context.Context, id, userID string) error
}
