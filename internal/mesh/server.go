package mesh

import (
	"context"
	"fmt"
	"log"

	"github.com/zoobz-io/aegis"
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

// Compile-time assertions.
var (
	_ Identity = (*Server)(nil)
	_ Sessions = (*Server)(nil)
	_ Users    = (*Server)(nil)
	_ Tenants  = (*Server)(nil)
)

// Server implements the mesh service contracts using janus stores.
type Server struct {
	tenants          *stores.Tenants
	users            *stores.Users
	memberships      *stores.Memberships
	linkedIdentities *stores.LinkedIdentities
	sessions         *stores.Sessions
}

// NewServer creates a new mesh server.
func NewServer(tenants *stores.Tenants, users *stores.Users, memberships *stores.Memberships, linkedIdentities *stores.LinkedIdentities, sessions *stores.Sessions) *Server {
	return &Server{
		tenants:          tenants,
		users:            users,
		memberships:      memberships,
		linkedIdentities: linkedIdentities,
		sessions:         sessions,
	}
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

// ResolveIdentity looks up an internal user by external IdP subject.
// Users must be registered first — returns an error if not found.
func (s *Server) ResolveIdentity(ctx context.Context, req ResolveIdentityRequest) (*ResolveIdentityResponse, error) {
	link, err := s.linkedIdentities.GetByProviderSubject(ctx, req.Provider, req.ExternalSubject)
	if err != nil {
		return nil, fmt.Errorf("looking up linked identity: %w", err)
	}
	if link == nil {
		return nil, fmt.Errorf("user not registered: no linked identity for %s/%s", req.Provider, req.ExternalSubject)
	}

	user, err := s.users.GetUser(ctx, link.UserID)
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	if touchErr := s.users.TouchLastSeen(ctx, user.ID); touchErr != nil {
		log.Printf("mesh: failed to touch last_seen for user %s: %v", user.ID, touchErr)
	}

	memberships, err := s.loadMemberships(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &ResolveIdentityResponse{
		UserID:      user.ID,
		Email:       user.Email,
		Memberships: memberships,
	}, nil
}

// Register creates a tenant, user, owner membership, and linked identity in one operation.
// Used by the registration flow.
// Register creates a user and links their external identity.
// If tenantName and tenantSlug are provided, also creates a tenant and
// owner membership. Otherwise the user is created without a tenant
// (they can create one later via POST /me/tenants).
func (s *Server) Register(ctx context.Context, tenantName, tenantSlug, email, displayName, provider, externalSubject string) (*ResolveIdentityResponse, error) {
	user, err := s.users.CreateUser(ctx, email, displayName)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	if _, err := s.linkedIdentities.Link(ctx, user.ID, provider, externalSubject); err != nil {
		return nil, fmt.Errorf("linking identity: %w", err)
	}

	capitan.Emit(ctx, events.UserCreated, events.UserIDKey.Field(user.ID), events.EmailKey.Field(email))
	capitan.Emit(ctx, events.IdentityLinked, events.UserIDKey.Field(user.ID), events.ProviderKey.Field(provider))

	resp := &ResolveIdentityResponse{
		UserID: user.ID,
		Email:  user.Email,
	}

	// Optionally create tenant + owner membership.
	if tenantName != "" && tenantSlug != "" {
		tenant, err := s.tenants.CreateTenant(ctx, tenantName, tenantSlug)
		if err != nil {
			return nil, fmt.Errorf("creating tenant: %w", err)
		}

		if _, err := s.memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleOwner); err != nil {
			return nil, fmt.Errorf("creating owner membership: %w", err)
		}

		capitan.Emit(ctx, events.TenantCreated, events.TenantIDKey.Field(tenant.ID))
		resp.Memberships = []MembershipInfo{{TenantID: tenant.ID, Role: models.UserRoleOwner}}
	}

	return resp, nil
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

// CreateSession issues a new session token for a user.
// The IssuedBy field is populated from the calling service's cert CN.
func (s *Server) CreateSession(ctx context.Context, req CreateSessionRequest) (*CreateSessionResponse, error) {
	issuedBy := req.IssuedBy
	if issuedBy == "" {
		if caller, err := aegis.CallerFromContext(ctx); err == nil {
			issuedBy = caller.NodeID
		} else {
			issuedBy = "unknown"
		}
	}

	token, sess, err := s.sessions.CreateSession(ctx, req.UserID, issuedBy, req.UserAgent, req.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	capitan.Emit(ctx, events.SessionCreated,
		events.SessionIDKey.Field(sess.ID),
		events.UserIDKey.Field(req.UserID),
		events.IssuedByKey.Field(issuedBy),
	)

	return &CreateSessionResponse{
		Token:     token,
		ExpiresAt: sess.ExpiresAt.Unix(),
	}, nil
}

// ValidateSession checks if a token is valid and returns the associated identity.
func (s *Server) ValidateSession(ctx context.Context, token string) (*ValidateSessionResponse, error) {
	sess, err := s.sessions.ValidateByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("validating session: %w", err)
	}
	if sess == nil {
		return &ValidateSessionResponse{Valid: false}, nil
	}

	user, err := s.users.GetUser(ctx, sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("looking up session user: %w", err)
	}

	memberships, err := s.loadMemberships(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &ValidateSessionResponse{
		Valid:       true,
		UserID:      user.ID,
		Email:       user.Email,
		IssuedBy:    sess.IssuedBy,
		Memberships: memberships,
	}, nil
}

// RevokeSession revokes a specific session by token.
func (s *Server) RevokeSession(ctx context.Context, token string) error {
	if err := s.sessions.RevokeByToken(ctx, token); err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	capitan.Emit(ctx, events.SessionRevoked)
	return nil
}

// RevokeUserSessions revokes all sessions for a user.
func (s *Server) RevokeUserSessions(ctx context.Context, userID string) (int, error) {
	count, err := s.sessions.RevokeUserSessions(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("revoking user sessions: %w", err)
	}
	capitan.Emit(ctx, events.SessionRevoked, events.UserIDKey.Field(userID))
	return count, nil
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

// GetUser retrieves a user by internal ID.
func (s *Server) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.users.GetUser(ctx, id)
}

// GetUserByEmail retrieves a user by email address.
func (s *Server) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.users.GetUserByEmail(ctx, email)
}

// ---------------------------------------------------------------------------
// Tenants
// ---------------------------------------------------------------------------

// GetTenant retrieves a tenant by ID.
func (s *Server) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	return s.tenants.GetTenant(ctx, id)
}

// CreateTenant provisions a new tenant.
func (s *Server) CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error) {
	tenant, err := s.tenants.CreateTenant(ctx, name, slug)
	if err != nil {
		return nil, err
	}
	capitan.Emit(ctx, events.TenantCreated, events.TenantIDKey.Field(tenant.ID))
	return tenant, nil
}

// UpdateTenant updates a tenant's name and status.
func (s *Server) UpdateTenant(ctx context.Context, id, name string, status models.TenantStatus) (*models.Tenant, error) {
	tenant, err := s.tenants.UpdateTenant(ctx, id, name, status)
	if err != nil {
		return nil, err
	}
	capitan.Emit(ctx, events.TenantUpdated, events.TenantIDKey.Field(tenant.ID))
	return tenant, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (s *Server) loadMemberships(ctx context.Context, userID string) ([]MembershipInfo, error) {
	mems, err := s.memberships.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("loading memberships: %w", err)
	}
	result := make([]MembershipInfo, len(mems))
	for i, m := range mems {
		result[i] = MembershipInfo{TenantID: m.TenantID, Role: m.Role}
	}
	return result, nil
}
