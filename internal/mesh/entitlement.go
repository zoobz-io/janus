package mesh

import (
	"context"
	"fmt"

	"github.com/zoobz-io/aegis"
	commonpb "github.com/zoobz-io/aegis/proto/common/v1"

	"github.com/zoobz-io/janus/stores"
)

// entitlementChecker resolves the calling application from the security context
// and verifies user entitlements (TenantApplication + UserApplication).
type entitlementChecker struct {
	applications       *stores.Applications
	tenantApplications *stores.TenantApplications
	userApplications   *stores.UserApplications
	memberships        *stores.Memberships
	tenants            *stores.Tenants
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
		ta, err := e.tenantApplications.GetByTenantAndApp(ctx, mem.TenantID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("checking tenant application: %w", err)
		}
		if ta == nil {
			continue
		}

		ua, err := e.userApplications.GetByUserAndApp(ctx, userID, mem.TenantID, app.ID)
		if err != nil {
			return nil, fmt.Errorf("checking user application: %w", err)
		}
		if ua == nil {
			continue
		}

		tenant, err := e.tenants.GetTenant(ctx, mem.TenantID)
		if err != nil {
			return nil, fmt.Errorf("looking up tenant: %w", err)
		}

		result = append(result, &commonpb.AuthorizedTenant{
			TenantId:   tenant.ID,
			TenantName: tenant.Name,
			Role:       mem.Role,
			AppRoles:   ua.Roles,
			AppScopes:  ua.Scopes,
		})
	}

	return result, nil
}
