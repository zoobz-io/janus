// Package mesh defines the internal service contracts exposed over the aegis mesh.
// These are implemented by janus stores and consumed by mesh-connected apps
// via gRPC. The proto definitions in aegis will mirror these interfaces.
package mesh

import (
	"context"

	"github.com/zoobz-io/janus/models"
)

// Identity defines operations for resolving external IdP identities
// to internal users. This is the core flow: an app authenticates a user
// against its own IdP, then calls janus to resolve the internal identity.
type Identity interface {
	// ResolveIdentity maps an external IdP subject to an internal user.
	// If the user does not exist, it is created (JIT provisioning).
	// If the tenant does not exist, returns an error — tenants must be
	// created explicitly via the Tenants interface.
	ResolveIdentity(ctx context.Context, req ResolveIdentityRequest) (*ResolveIdentityResponse, error)
}

// ResolveIdentityRequest contains the external identity to resolve.
type ResolveIdentityRequest struct {
	Provider        string // IdP type: "zitadel", "auth0", "github", "google"
	ExternalSubject string // Subject ID from the IdP
	TenantID        string // Internal tenant the user belongs to
	Email           string // User email from IdP claims
	DisplayName     string // User display name from IdP claims
}

// ResolveIdentityResponse contains the resolved internal identity.
type ResolveIdentityResponse struct {
	UserID   string // Internal user ID
	TenantID string // Internal tenant ID
	Created  bool   // True if the user was just provisioned
}

// Sessions defines operations for managing mesh-wide session tokens.
type Sessions interface {
	// CreateSession issues a new session token for a user.
	// The issuedBy field is populated from the calling service's identity.
	CreateSession(ctx context.Context, req CreateSessionRequest) (*CreateSessionResponse, error)
	// ValidateSession checks if a token is valid and returns the associated identity.
	ValidateSession(ctx context.Context, token string) (*ValidateSessionResponse, error)
	// RevokeSession revokes a specific session by token.
	RevokeSession(ctx context.Context, token string) error
	// RevokeUserSessions revokes all sessions for a user.
	RevokeUserSessions(ctx context.Context, userID string) (int, error)
}

// CreateSessionRequest contains the parameters for creating a session.
type CreateSessionRequest struct {
	UserID    string // Internal user ID
	IssuedBy  string // Node ID of the calling service (from cert CN)
	UserAgent string // Device/browser info
	IPAddress string // Client IP
}

// CreateSessionResponse contains the issued session token.
type CreateSessionResponse struct {
	Token     string // Opaque session token — only returned once
	ExpiresAt int64  // Unix timestamp
}

// ValidateSessionResponse contains the validated session identity.
type ValidateSessionResponse struct {
	Valid    bool
	UserID   string
	TenantID string
	IssuedBy string
	Email    string
	Role     string
}

// Users defines read operations for user queries over the mesh.
type Users interface {
	// GetUser retrieves a user by internal ID.
	GetUser(ctx context.Context, id string) (*models.User, error)
	// GetUserByEmail retrieves a user by email address.
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	// ListUsersByTenant retrieves a paginated list of users for a tenant.
	ListUsersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.User], error)
}

// Tenants defines operations for tenant management over the mesh.
type Tenants interface {
	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
	// CreateTenant provisions a new tenant.
	CreateTenant(ctx context.Context, name, slug string) (*models.Tenant, error)
	// UpdateTenant updates a tenant's name and status.
	UpdateTenant(ctx context.Context, id, name string, status models.TenantStatus) (*models.Tenant, error)
}
