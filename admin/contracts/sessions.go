package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Sessions defines the admin API's capability boundary over the sessions store.
type Sessions interface {
	ListSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error)
	RevokeSession(ctx context.Context, id, userID string) error
	RevokeUserSessions(ctx context.Context, userID string) (int, error)
}
