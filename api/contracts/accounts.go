package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Accounts defines the contract for account operations on the public API surface.
type Accounts interface {
	// ListByUser retrieves all accounts for a user.
	ListByUser(ctx context.Context, userID string) ([]*models.Account, error)
	// Unlink removes a account by ID, scoped to the user.
	Unlink(ctx context.Context, id, userID string) error
}
