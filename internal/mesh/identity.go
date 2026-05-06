package mesh

import (
	"context"
	"fmt"
	"log"

	"github.com/zoobz-io/capitan"
	identitypb "github.com/zoobz-io/aegis/proto/identity/v1"

	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

// IdentityServer implements identitypb.IdentityServiceServer.
type IdentityServer struct {
	identitypb.UnimplementedIdentityServiceServer
	users            *stores.Users
	linkedIdentities *stores.LinkedIdentities
	tenants          *stores.Tenants
	memberships      *stores.Memberships
	entitlement      *entitlementChecker
}

// NewIdentityServer creates a new IdentityServer.
func NewIdentityServer(
	users *stores.Users,
	linkedIdentities *stores.LinkedIdentities,
	tenants *stores.Tenants,
	memberships *stores.Memberships,
	applications *stores.Applications,
	tenantApplications *stores.TenantApplications,
	userApplications *stores.UserApplications,
) *IdentityServer {
	return &IdentityServer{
		users:            users,
		linkedIdentities: linkedIdentities,
		tenants:          tenants,
		memberships:      memberships,
		entitlement: &entitlementChecker{
			applications:       applications,
			tenantApplications: tenantApplications,
			userApplications:   userApplications,
			memberships:        memberships,
			tenants:            tenants,
		},
	}
}

// ResolveIdentity looks up an internal user by external IdP provider and subject.
// The response is scoped to the calling application — if the user does not
// have an entitlement for the caller's app, an error is returned.
func (s *IdentityServer) ResolveIdentity(ctx context.Context, req *identitypb.ResolveIdentityRequest) (*identitypb.ResolveIdentityResponse, error) {
	link, err := s.linkedIdentities.GetByProviderSubject(ctx, req.Provider, req.ProviderUserId)
	if err != nil {
		return nil, fmt.Errorf("looking up linked identity: %w", err)
	}
	if link == nil {
		return nil, fmt.Errorf("user not registered: no linked identity for %s/%s", req.Provider, req.ProviderUserId)
	}

	user, err := s.users.GetUser(ctx, link.UserID)
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	if touchErr := s.users.TouchLastSeen(ctx, user.ID); touchErr != nil {
		log.Printf("mesh: failed to touch last_seen for user %s: %v", user.ID, touchErr)
	}

	// Check entitlement for the calling application.
	appSlug, err := callerAppSlug(ctx)
	if err != nil {
		log.Printf("mesh: no caller identity for entitlement check: %v", err)
		return &identitypb.ResolveIdentityResponse{
			UserId:  user.ID,
			Created: false,
		}, nil
	}

	tenants, err := s.entitlement.authorizedTenants(ctx, user.ID, appSlug)
	if err != nil {
		return nil, fmt.Errorf("checking entitlement: %w", err)
	}
	if len(tenants) == 0 {
		return nil, fmt.Errorf("user %s is not entitled to application %s", user.ID, appSlug)
	}

	return &identitypb.ResolveIdentityResponse{
		UserId:  user.ID,
		Created: false,
		Tenants: tenants,
	}, nil
}

// Register creates a user, links an external identity, and optionally creates a tenant.
func (s *IdentityServer) Register(ctx context.Context, req *identitypb.RegisterRequest) (*identitypb.RegisterResponse, error) {
	user, err := s.users.CreateUser(ctx, req.Email, req.Name)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	if _, err := s.linkedIdentities.Link(ctx, user.ID, req.Provider, req.ProviderUserId); err != nil {
		return nil, fmt.Errorf("linking identity: %w", err)
	}

	capitan.Emit(ctx, events.UserCreated, events.UserIDKey.Field(user.ID), events.EmailKey.Field(req.Email))
	capitan.Emit(ctx, events.IdentityLinked, events.UserIDKey.Field(user.ID), events.ProviderKey.Field(req.Provider))

	resp := &identitypb.RegisterResponse{
		UserId: user.ID,
	}

	if req.TenantName != "" && req.TenantSlug != "" {
		tenant, err := s.tenants.CreateTenant(ctx, req.TenantName, req.TenantSlug)
		if err != nil {
			return nil, fmt.Errorf("creating tenant: %w", err)
		}

		if _, err := s.memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleOwner); err != nil {
			return nil, fmt.Errorf("creating owner membership: %w", err)
		}

		capitan.Emit(ctx, events.TenantCreated, events.TenantIDKey.Field(tenant.ID))
		resp.TenantId = tenant.ID
	}

	return resp, nil
}

// ListProviders returns the external OAuth providers linked to a user.
func (s *IdentityServer) ListProviders(ctx context.Context, req *identitypb.ListProvidersRequest) (*identitypb.ListProvidersResponse, error) {
	identities, err := s.linkedIdentities.ListByUser(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("listing linked identities: %w", err)
	}

	providers := make([]*identitypb.Provider, len(identities))
	for i, id := range identities {
		providers[i] = &identitypb.Provider{
			Provider:       id.Provider,
			ProviderUserId: id.ExternalSubject,
			LinkedAt:       id.LinkedAt.Unix(),
		}
	}

	return &identitypb.ListProvidersResponse{
		Providers: providers,
	}, nil
}
