package mesh

import (
	"context"
	"fmt"

	"github.com/zoobz-io/aegis"
	commonpb "github.com/zoobz-io/aegis/proto/common/v1"

	"github.com/zoobz-io/janus/internal/authz"
	"github.com/zoobz-io/janus/stores"
)

// entitlementChecker adapts the shared authz entitlement resolver to the
// mesh's protobuf surface: the calling application is identified from the
// security context (cert CN) and resolved tenants are mapped to
// commonpb.AuthorizedTenant.
type entitlementChecker struct {
	entitlements *authz.Entitlements
}

// newEntitlementChecker wires the shared resolver over the concrete stores.
func newEntitlementChecker(
	applications *stores.Applications,
	licenses *stores.Licenses,
	grants *stores.Grants,
	memberships *stores.Memberships,
	tenants *stores.Tenants,
	features *stores.Features,
	scopes *stores.Scopes,
) *entitlementChecker {
	return &entitlementChecker{
		entitlements: authz.NewEntitlements(
			applications, licenses, grants, memberships, tenants, features, scopes,
		),
	}
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
	_, tenants, err := e.entitlements.ForApplication(ctx, userID, appSlug)
	if err != nil {
		return nil, err
	}

	result := make([]*commonpb.AuthorizedTenant, len(tenants))
	for i, t := range tenants {
		result[i] = &commonpb.AuthorizedTenant{
			TenantId:   t.TenantID,
			TenantName: t.TenantName,
			Role:       t.Role,
			AppRoles:   t.AppRoles,
			AppScopes:  t.AppScopes,
		}
	}
	return result, nil
}
