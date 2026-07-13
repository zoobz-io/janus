package mesh

import (
	"context"
	"fmt"

	"github.com/zoobz-io/aegis"
	sessionpb "github.com/zoobz-io/aegis/proto/session/v1"
	"github.com/zoobz-io/capitan"
	"google.golang.org/grpc"

	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/stores"
)

// SessionServer implements sessionpb.SessionServiceServer.
type SessionServer struct {
	sessionpb.UnimplementedSessionServiceServer
	sessions    *stores.Sessions
	users       *stores.Users
	entitlement *entitlementChecker
}

// NewSessionServer creates a new SessionServer.
func NewSessionServer(
	sessions *stores.Sessions,
	users *stores.Users,
	tenants *stores.Tenants,
	applications *stores.Applications,
	licenses *stores.Licenses,
	grants *stores.Grants,
	memberships *stores.Memberships,
	features *stores.Features,
	scopes *stores.Scopes,
) *SessionServer {
	return &SessionServer{
		sessions: sessions,
		users:    users,
		entitlement: &entitlementChecker{
			applications: applications,
			licenses:     licenses,
			grants:       grants,
			memberships:  memberships,
			tenants:      tenants,
			features:     features,
			scopes:       scopes,
		},
	}
}

// CreateSession issues a session token for a user.
func (s *SessionServer) CreateSession(ctx context.Context, req *sessionpb.CreateSessionRequest) (*sessionpb.CreateSessionResponse, error) {
	issuedBy := "unknown"
	if caller, err := aegis.CallerFromContext(ctx); err == nil {
		issuedBy = caller.NodeID
	}

	token, sess, err := s.sessions.CreateSession(ctx, req.UserId, issuedBy, req.UserAgent, req.IpAddress)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	events.SessionCreated.Emit(ctx, events.SessionEvent{
		SessionID: sess.ID,
		UserID:    req.UserId,
		IssuedBy:  issuedBy,
	})

	return &sessionpb.CreateSessionResponse{
		Token:     token,
		ExpiresAt: sess.ExpiresAt.Unix(),
	}, nil
}

// ValidateSession checks if a session token is valid and returns user info.
// The response is scoped to the calling application — if the user does not
// have an entitlement for the caller's app, valid is false.
func (s *SessionServer) ValidateSession(ctx context.Context, req *sessionpb.ValidateSessionRequest) (*sessionpb.ValidateSessionResponse, error) {
	sess, err := s.sessions.ValidateByToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("validating session: %w", err)
	}
	if sess == nil {
		return &sessionpb.ValidateSessionResponse{Valid: false}, nil
	}

	// Check entitlement for the calling application.
	appSlug, err := callerAppSlug(ctx)
	if err != nil {
		capitan.Warn(ctx, events.EntitlementCheckSkipped, events.OpErrorKey.Field(err))
		return &sessionpb.ValidateSessionResponse{
			Valid:     true,
			UserId:    sess.UserID,
			ExpiresAt: sess.ExpiresAt.Unix(),
		}, nil
	}

	tenants, err := s.entitlement.authorizedTenants(ctx, sess.UserID, appSlug)
	if err != nil {
		return nil, fmt.Errorf("checking entitlement: %w", err)
	}
	if len(tenants) == 0 {
		return &sessionpb.ValidateSessionResponse{Valid: false}, nil
	}

	return &sessionpb.ValidateSessionResponse{
		Valid:     true,
		UserId:    sess.UserID,
		ExpiresAt: sess.ExpiresAt.Unix(),
		Tenants:   tenants,
	}, nil
}

// RevokeSession revokes a specific session by token.
func (s *SessionServer) RevokeSession(ctx context.Context, req *sessionpb.RevokeSessionRequest) (*sessionpb.RevokeSessionResponse, error) {
	if err := s.sessions.RevokeByToken(ctx, req.Token); err != nil {
		return nil, fmt.Errorf("revoking session: %w", err)
	}
	events.SessionRevoked.Emit(ctx, events.SessionEvent{})
	return &sessionpb.RevokeSessionResponse{}, nil
}

// RevokeUserSessions revokes all sessions for a user.
func (s *SessionServer) RevokeUserSessions(ctx context.Context, req *sessionpb.RevokeUserSessionsRequest) (*sessionpb.RevokeUserSessionsResponse, error) {
	count, err := s.sessions.RevokeUserSessions(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("revoking user sessions: %w", err)
	}
	events.SessionRevoked.Emit(ctx, events.SessionEvent{UserID: req.UserId})
	return &sessionpb.RevokeUserSessionsResponse{
		RevokedCount: uint32(count), //nolint:gosec // count is bounded by session table rows
	}, nil
}

// ListUserSessions lists active sessions for a user.
func (s *SessionServer) ListUserSessions(ctx context.Context, req *sessionpb.ListUserSessionsRequest) (*sessionpb.ListUserSessionsResponse, error) {
	sessions, err := s.sessions.ListSessionsByUser(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	infos := make([]*sessionpb.SessionInfo, len(sessions))
	for i, sess := range sessions {
		prefix := sess.ID
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		infos[i] = &sessionpb.SessionInfo{
			TokenPrefix: prefix,
			CreatedAt:   sess.CreatedAt.Unix(),
			ExpiresAt:   sess.ExpiresAt.Unix(),
			UserAgent:   sess.UserAgent,
			IpAddress:   sess.IPAddress,
		}
	}

	return &sessionpb.ListUserSessionsResponse{
		Sessions: infos,
	}, nil
}

// SubscribeSessionEvents is not yet implemented.
func (s *SessionServer) SubscribeSessionEvents(_ *sessionpb.SubscribeSessionEventsRequest, _ grpc.ServerStreamingServer[sessionpb.SessionEvent]) error {
	return fmt.Errorf("session event streaming not yet implemented")
}
