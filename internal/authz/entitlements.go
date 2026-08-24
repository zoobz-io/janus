package authz

import (
	"context"
	"fmt"
	"sort"

	"github.com/zoobz-io/janus/database/models"
)

// The store access entitlement resolution needs, as small structural
// interfaces so authz stays decoupled from any single API's contract package
// (the concrete stores satisfy all of them).

// ApplicationLookup resolves applications by slug.
type ApplicationLookup interface {
	GetBySlug(ctx context.Context, slug string) (*models.Application, error)
}

// LicenseLookup checks tenant-level application authorization.
type LicenseLookup interface {
	GetByTenantAndApp(ctx context.Context, tenantID, applicationID string) (*models.License, error)
}

// GrantLookup checks user-level application access within a tenant.
type GrantLookup interface {
	GetByUserAndApp(ctx context.Context, userID, tenantID, applicationID string) (*models.Grant, error)
}

// MembershipLookup lists a user's tenant memberships.
type MembershipLookup interface {
	ListByUser(ctx context.Context, userID string) ([]*models.Membership, error)
}

// TenantLookup resolves tenants by ID.
type TenantLookup interface {
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
}

// UserLookup resolves users by ID — needed to honor account status.
type UserLookup interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
}

// FeatureLookup lists the features bundled into a tier.
type FeatureLookup interface {
	ListByTier(ctx context.Context, tierID string) ([]*models.Feature, error)
}

// ScopeLookup resolves scopes by ID.
type ScopeLookup interface {
	GetByID(ctx context.Context, id string) (*models.Scope, error)
}

// Entitlements resolves a user's authorization for an application from the
// licensing and grant data: which tenants entitle the user to the app, with
// what role, application roles, and effective scopes. It is the one
// resolution path shared by the mesh (session validation, identity) and the
// public HTTP API.
type Entitlements struct {
	applications ApplicationLookup
	licenses     LicenseLookup
	grants       GrantLookup
	memberships  MembershipLookup
	tenants      TenantLookup
	features     FeatureLookup
	scopes       ScopeLookup
	users        UserLookup
}

// NewEntitlements creates an entitlement resolver over the given stores.
func NewEntitlements(
	applications ApplicationLookup,
	licenses LicenseLookup,
	grants GrantLookup,
	memberships MembershipLookup,
	tenants TenantLookup,
	features FeatureLookup,
	scopes ScopeLookup,
	users UserLookup,
) *Entitlements {
	return &Entitlements{
		applications: applications,
		licenses:     licenses,
		grants:       grants,
		memberships:  memberships,
		tenants:      tenants,
		features:     features,
		scopes:       scopes,
		users:        users,
	}
}

// ForApplication resolves the application by slug and the tenants through
// which the user is entitled to it, with their role in each. The tenant list
// is empty (not an error) when the user has no entitlements; an unknown slug
// returns an error wrapping ErrApplicationNotFound.
func (e *Entitlements) ForApplication(ctx context.Context, userID, appSlug string) (*models.Application, []models.AuthorizedTenant, error) {
	app, err := e.applications.GetBySlug(ctx, appSlug)
	if err != nil || app == nil {
		return nil, nil, fmt.Errorf("looking up application %q: %w", appSlug, ErrApplicationNotFound)
	}

	// A deactivated user has no entitlements, regardless of grants — access is
	// lost on the next request, not at token expiry.
	user, err := e.users.GetUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("looking up user: %w", err)
	}
	if user.Status != models.UserStatusActive {
		return app, nil, nil
	}

	mems, err := e.memberships.ListByUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing user memberships: %w", err)
	}

	result := make([]models.AuthorizedTenant, 0, len(mems))
	for _, mem := range mems {
		license, err := e.licenses.GetByTenantAndApp(ctx, mem.TenantID, app.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("checking license: %w", err)
		}
		if license == nil {
			continue
		}

		grant, err := e.grants.GetByUserAndApp(ctx, userID, mem.TenantID, app.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("checking grant: %w", err)
		}
		if grant == nil {
			continue
		}

		tenant, err := e.tenants.GetTenant(ctx, mem.TenantID)
		if err != nil {
			return nil, nil, fmt.Errorf("looking up tenant: %w", err)
		}
		// A suspended tenant entitles no one, even members with live grants.
		if tenant.Status != models.TenantStatusActive {
			continue
		}

		scopes, err := e.effectiveScopes(ctx, grant)
		if err != nil {
			return nil, nil, err
		}

		result = append(result, models.AuthorizedTenant{
			TenantID:   tenant.ID,
			TenantName: tenant.Name,
			Role:       mem.Role,
			AppRoles:   grant.Roles,
			AppScopes:  scopes,
		})
	}

	return app, result, nil
}

// effectiveScopes composes the scopes granted to a user for an application:
// the explicit scopes assigned directly on the grant (admin-assigned) unioned
// with the scopes inherited from the grant's tier (its bundled features).
// Resolved dynamically so tier edits take effect immediately.
func (e *Entitlements) effectiveScopes(ctx context.Context, g *models.Grant) ([]string, error) {
	set := make(map[string]struct{}, len(g.Scopes))
	for _, s := range g.Scopes {
		set[s] = struct{}{}
	}

	if g.TierID != nil {
		feats, err := e.features.ListByTier(ctx, *g.TierID)
		if err != nil {
			return nil, fmt.Errorf("listing tier features: %w", err)
		}
		for _, f := range feats {
			sc, err := e.scopes.GetByID(ctx, f.ScopeID)
			if err != nil {
				return nil, fmt.Errorf("looking up scope %q: %w", f.ScopeID, err)
			}
			if sc != nil {
				set[sc.Name] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out, nil
}
