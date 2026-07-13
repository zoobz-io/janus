package mesh

import (
	"context"
	"fmt"
	"sort"

	"github.com/zoobz-io/aegis"
	commonpb "github.com/zoobz-io/aegis/proto/common/v1"

	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

// entitlementChecker resolves the calling application from the security context
// and verifies user entitlements (License + Grant).
type entitlementChecker struct {
	applications *stores.Applications
	licenses     *stores.Licenses
	grants       *stores.Grants
	memberships  *stores.Memberships
	tenants      *stores.Tenants
	features     *stores.Features
	scopes       *stores.Scopes
}

// callerAppSlug extracts the calling application's slug from the security context.
// The slug is the cert CN, which maps to Application.Slug.
func callerAppSlug(ctx context.Context) (string, error) {
	sc, ok := aegis.SecurityContextFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no security context available")
	}
	if sc.Metadata.NodeID == "" {
		return "", fmt.Errorf("caller has no node ID")
	}
	return sc.Metadata.NodeID, nil
}

// authorizedTenants returns the tenants through which the user is entitled to
// the given application, with their role in each. Returns nil if the user has
// no entitlements.
func (e *entitlementChecker) authorizedTenants(ctx context.Context, userID, appSlug string) ([]*commonpb.AuthorizedTenant, error) {
	app, err := e.applications.GetBySlug(ctx, appSlug)
	if err != nil {
		return nil, fmt.Errorf("looking up application %q: %w", appSlug, err)
	}

	mems, err := e.memberships.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user memberships: %w", err)
	}

	result := make([]*commonpb.AuthorizedTenant, 0, len(mems))
	for _, mem := range mems {
		ta, err := e.licenses.GetByTenantAndApp(ctx, mem.TenantID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("checking license: %w", err)
		}
		if ta == nil {
			continue
		}

		ua, err := e.grants.GetByUserAndApp(ctx, userID, mem.TenantID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("checking grant: %w", err)
		}
		if ua == nil {
			continue
		}

		tenant, err := e.tenants.GetTenant(ctx, mem.TenantID)
		if err != nil {
			return nil, fmt.Errorf("looking up tenant: %w", err)
		}

		scopes, err := e.effectiveScopes(ctx, ua)
		if err != nil {
			return nil, err
		}

		result = append(result, &commonpb.AuthorizedTenant{
			TenantId:   tenant.ID,
			TenantName: tenant.Name,
			Role:       mem.Role,
			AppRoles:   ua.Roles,
			AppScopes:  scopes,
		})
	}

	return result, nil
}

// effectiveScopes composes the scopes granted to a user for an application:
// the explicit scopes assigned directly on the grant (admin-assigned) unioned
// with the scopes inherited from the grant's tier (its bundled features).
// Resolved dynamically so tier edits take effect immediately.
func (e *entitlementChecker) effectiveScopes(ctx context.Context, g *models.Grant) ([]string, error) {
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
