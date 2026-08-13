package contracts

import (
	"context"

	"github.com/zoobz-io/janus/database/models"
)

// Sessions defines the contract for session operations on the public API surface.
type Sessions interface {
	// ListSessionsByUser retrieves active sessions for a user.
	ListSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error)
	// RevokeSession revokes a specific session by ID, scoped to the user.
	RevokeSession(ctx context.Context, id, userID string) error
	// RevokeUserSessions revokes all sessions for a user.
	RevokeUserSessions(ctx context.Context, userID string) (int, error)
}
