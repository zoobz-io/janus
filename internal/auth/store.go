package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zoobz-io/rocco/session"

	"github.com/zoobz-io/janus/database/stores"
)

// Compile-time assertion.
var _ session.Store = (*SessionStore)(nil)

// stateTTL is how long CSRF state tokens live in Redis.
const stateTTL = 10 * time.Minute

// SessionStore implements rocco/session.Store backed by the janus sessions
// table (for sessions) and Redis (for CSRF state tokens).
type SessionStore struct {
	sessions *stores.Sessions
	users    *stores.Users
	redis    *goredis.Client
}

// NewSessionStore creates a new session store adapter.
func NewSessionStore(sessions *stores.Sessions, users *stores.Users, redis *goredis.Client) *SessionStore {
	return &SessionStore{
		sessions: sessions,
		users:    users,
		redis:    redis,
	}
}

// CreateState persists a CSRF state token in Redis with a TTL.
func (s *SessionStore) CreateState(ctx context.Context, state string) error {
	return s.redis.Set(ctx, stateKey(state), "1", stateTTL).Err()
}

// VerifyState checks that the state token exists and deletes it (single-use).
func (s *SessionStore) VerifyState(ctx context.Context, state string) (bool, error) {
	key := stateKey(state)
	result, err := s.redis.GetDel(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("verifying state: %w", err)
	}
	return result == "1", nil
}

// Create persists a new session.
func (s *SessionStore) Create(ctx context.Context, id string, data session.Data) error {
	_, _, err := s.sessions.CreateSessionWithID(ctx, id, data.UserID, "janus", "", "")
	return err
}

// Get retrieves session data by ID.
func (s *SessionStore) Get(ctx context.Context, id string) (*session.Data, error) {
	sess, err := s.sessions.GetSession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	if sess.Expired() {
		return nil, fmt.Errorf("session expired")
	}

	user, err := s.users.GetUser(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("user lookup failed: %w", err)
	}

	return &session.Data{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

// Refresh is a no-op — session expiry is managed by the sessions store.
func (s *SessionStore) Refresh(_ context.Context, _ string) error {
	return nil
}

// Delete removes a session by ID.
func (s *SessionStore) Delete(ctx context.Context, id string) error {
	return s.sessions.DeleteSession(ctx, id)
}

func stateKey(state string) string {
	return "janus:oauth:state:" + state
}
