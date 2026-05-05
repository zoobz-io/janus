package stores

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/models"
)

// DefaultSessionDuration is the default session lifetime.
const DefaultSessionDuration = 24 * time.Hour

// Sessions provides database access for sessions.
type Sessions struct {
	*sum.Database[models.Session]
}

// NewSessions creates a new sessions store.
func NewSessions(db *sqlx.DB, renderer astql.Renderer) *Sessions {
	return &Sessions{
		Database: sum.NewDatabase[models.Session](db, "sessions", renderer),
	}
}

// CreateSession creates a new session and returns the raw token.
// The raw token is only returned once — only the hash is stored.
func (s *Sessions) CreateSession(ctx context.Context, userID, issuedBy, userAgent, ipAddress string) (string, *models.Session, error) {
	token, err := generateToken()
	if err != nil {
		return "", nil, fmt.Errorf("generating session token: %w", err)
	}

	sess := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: hashToken(token),
		IssuedBy:  issuedBy,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(DefaultSessionDuration),
	}
	if err := s.Set(ctx, "", sess); err != nil {
		return "", nil, fmt.Errorf("creating session: %w", err)
	}
	return token, sess, nil
}

// CreateSessionWithID creates a session with a specific ID (used by the OAuth cookie flow).
// The token hash is derived from the provided ID since cookie sessions don't use bearer tokens.
func (s *Sessions) CreateSessionWithID(ctx context.Context, id, userID, issuedBy, userAgent, ipAddress string) (string, *models.Session, error) {
	sess := &models.Session{
		ID:        id,
		UserID:    userID,
		TokenHash: hashToken(id),
		IssuedBy:  issuedBy,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: time.Now().Add(DefaultSessionDuration),
	}
	if err := s.Set(ctx, "", sess); err != nil {
		return "", nil, fmt.Errorf("creating session: %w", err)
	}
	return id, sess, nil
}

// GetSession retrieves a session by ID.
func (s *Sessions) GetSession(ctx context.Context, id string) (*models.Session, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// DeleteSession removes a session by ID (no user scope check).
func (s *Sessions) DeleteSession(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ValidateByToken looks up a session by token hash and checks expiry.
func (s *Sessions) ValidateByToken(ctx context.Context, token string) (*models.Session, error) {
	hash := hashToken(token)
	results, err := s.Query().
		Where("token_hash", "=", "token_hash").
		Limit(1).
		Exec(ctx, map[string]any{"token_hash": hash})
	if err != nil {
		return nil, fmt.Errorf("validating session: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}
	sess := results[0]
	if sess.Expired() {
		return nil, nil
	}
	return sess, nil
}

// ListSessionsByUser retrieves all active (non-expired) sessions for a user.
func (s *Sessions) ListSessionsByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	now := time.Now()
	return s.Query().
		Where("user_id", "=", "user_id").
		Where("expires_at", ">", "now").
		OrderBy("created_at", "DESC").
		Exec(ctx, map[string]any{"user_id": userID, "now": now})
}

// RevokeSession deletes a session by ID, scoped to a user.
func (s *Sessions) RevokeSession(ctx context.Context, id, userID string) error {
	results, err := s.Query().
		Where("id", "=", "id").
		Where("user_id", "=", "user_id").
		Limit(1).
		Exec(ctx, map[string]any{"id": id, "user_id": userID})
	if err != nil {
		return fmt.Errorf("finding session: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("session %w", ErrNotFound)
	}
	return s.Delete(ctx, id)
}

// RevokeByToken revokes a session by its raw token.
func (s *Sessions) RevokeByToken(ctx context.Context, token string) error {
	sess, err := s.ValidateByToken(ctx, token)
	if err != nil {
		return err
	}
	if sess == nil {
		return fmt.Errorf("session %w", ErrNotFound)
	}
	return s.Delete(ctx, sess.ID)
}

// RevokeUserSessions deletes all sessions for a user. Returns the count revoked.
func (s *Sessions) RevokeUserSessions(ctx context.Context, userID string) (int, error) {
	sessions, err := s.Query().
		Where("user_id", "=", "user_id").
		Exec(ctx, map[string]any{"user_id": userID})
	if err != nil {
		return 0, fmt.Errorf("listing sessions for revocation: %w", err)
	}
	for _, sess := range sessions {
		if err := s.Delete(ctx, sess.ID); err != nil {
			return 0, fmt.Errorf("revoking session %s: %w", sess.ID, err)
		}
	}
	return len(sessions), nil
}

// generateToken produces a cryptographically random 32-byte hex token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns the SHA-256 hex digest of a raw token.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
