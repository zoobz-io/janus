package contracts

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Accounts defines the admin API's capability boundary over the accounts store.
type Accounts interface {
	ListByUser(ctx context.Context, userID string) ([]*models.Account, error)
	Unlink(ctx context.Context, id, userID string) error
}
