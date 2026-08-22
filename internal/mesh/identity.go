package mesh

import (
	"context"
	"fmt"

	identitypb "github.com/zoobz-io/aegis/proto/identity/v1"
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/database/stores"
)

// IdentityServer implements identitypb.IdentityServiceServer.
type IdentityServer struct {
	identitypb.UnimplementedIdentityServiceServer
	stores      *stores.Stores
	users       *stores.Users
	accounts    *stores.Accounts
	entitlement *entitlementChecker
}

// NewIdentityServer creates a new IdentityServer over the store aggregate. The
// aggregate is required (not just individual stores) because registration is a
// transactional multi-store flow.
func NewIdentityServer(st *stores.Stores) *IdentityServer {
	return &IdentityServer{
		stores:   st,
		users:    st.Users,
		accounts: st.Accounts,
		entitlement: newEntitlementChecker(
			st.Applications, st.Licenses, st.Grants, st.Memberships, st.Tenants, st.Features, st.Scopes, st.Users,
		),
	}
}

// ResolveIdentity looks up an internal user by external IdP provider and subject.
func (s *IdentityServer) ResolveIdentity(ctx context.Context, req *identitypb.ResolveIdentityRequest) (*identitypb.ResolveIdentityResponse, error) {
	link, err := s.accounts.GetByProviderSubject(ctx, req.Provider, req.ProviderUserId)
	if err != nil {
		return nil, fmt.Errorf("looking up account: %w", err)
	}
	if link == nil {
		return nil, fmt.Errorf("user not registered: no account for %s/%s", req.Provider, req.ProviderUserId)
	}

	user, err := s.users.GetUser(ctx, link.UserID)
	if err != nil {
		return nil, fmt.Errorf("looking up user: %w", err)
	}

	if touchErr := s.users.TouchLastSeen(ctx, user.ID); touchErr != nil {
		capitan.Warn(ctx, events.LastSeenUpdateFailed, events.OpUserIDKey.Field(user.ID), events.OpErrorKey.Field(touchErr))
	}

	// Check entitlement for the calling application.
	appSlug, err := callerAppSlug(ctx)
	if err != nil {
		capitan.Warn(ctx, events.EntitlementCheckSkipped, events.OpErrorKey.Field(err))
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

// Register creates a user, links an external identity, and optionally creates
// a tenant — all in one transaction. Events emit only after the commit, so a
// rolled-back registration emits nothing.
func (s *IdentityServer) Register(ctx context.Context, req *identitypb.RegisterRequest) (*identitypb.RegisterResponse, error) {
	user, tenant, err := s.stores.RegisterUser(ctx, req.Email, req.Name, req.Provider, req.ProviderUserId, req.TenantName, req.TenantSlug)
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}

	events.UserCreated.Emit(ctx, events.UserEvent{UserID: user.ID, Email: req.Email})
	events.IdentityLinked.Emit(ctx, events.IdentityEvent{UserID: user.ID, Provider: req.Provider})

	resp := &identitypb.RegisterResponse{
		UserId: user.ID,
	}
	if tenant != nil {
		events.TenantCreated.Emit(ctx, events.TenantEvent{TenantID: tenant.ID})
		resp.TenantId = tenant.ID
	}
	return resp, nil
}

// ListProviders returns the external OAuth providers linked to a user.
func (s *IdentityServer) ListProviders(ctx context.Context, req *identitypb.ListProvidersRequest) (*identitypb.ListProvidersResponse, error) {
	identities, err := s.accounts.ListByUser(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
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
