package mesh

import (
	"context"
	"fmt"

	entitlementpb "github.com/zoobz-io/aegis/proto/entitlement/v1"

	"github.com/zoobz-io/janus/events"
	"github.com/zoobz-io/janus/internal/authz"
	"github.com/zoobz-io/janus/models"
	"github.com/zoobz-io/janus/stores"
)

// EntitlementServer implements entitlementpb.EntitlementServiceServer.
type EntitlementServer struct {
	entitlementpb.UnimplementedEntitlementServiceServer
	applications *stores.Applications
	licenses     *stores.Licenses
	grants       *stores.Grants
	memberships  *stores.Memberships
}

// NewEntitlementServer creates a new EntitlementServer.
func NewEntitlementServer(
	applications *stores.Applications,
	licenses *stores.Licenses,
	grants *stores.Grants,
	memberships *stores.Memberships,
) *EntitlementServer {
	return &EntitlementServer{
		applications: applications,
		licenses:     licenses,
		grants:       grants,
		memberships:  memberships,
	}
}

// resolveApp looks up the calling application by cert CN.
func (s *EntitlementServer) resolveApp(ctx context.Context) (*models.Application, error) {
	slug, err := callerAppSlug(ctx)
	if err != nil {
		return nil, fmt.Errorf("identifying calling application: %w", err)
	}
	app, err := s.applications.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("looking up application %q: %w", slug, err)
	}
	return app, nil
}

// requireAdmin verifies the acting user has admin or owner role in the tenant.
func (s *EntitlementServer) requireAdmin(ctx context.Context, actingUserID, tenantID string) error {
	_, err := authz.RequireRole(ctx, s.memberships, actingUserID, tenantID, models.UserRoleAdmin, models.UserRoleOwner)
	return err
}

// AuthorizeApplication grants a tenant access to the calling application.
func (s *EntitlementServer) AuthorizeApplication(ctx context.Context, req *entitlementpb.AuthorizeApplicationRequest) (*entitlementpb.AuthorizeApplicationResponse, error) {
	if err := s.requireAdmin(ctx, req.ActingUserId, req.TenantId); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	app, err := s.resolveApp(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := s.licenses.Authorize(ctx, req.TenantId, app.ID); err != nil {
		return nil, fmt.Errorf("authorizing application: %w", err)
	}

	events.TenantAppAuthorized.Emit(ctx, events.AppEntitlementEvent{TenantID: req.TenantId, ApplicationID: app.ID})
	return &entitlementpb.AuthorizeApplicationResponse{}, nil
}

// RevokeApplication revokes a tenant's access to the calling application.
func (s *EntitlementServer) RevokeApplication(ctx context.Context, req *entitlementpb.RevokeApplicationRequest) (*entitlementpb.RevokeApplicationResponse, error) {
	if err := s.requireAdmin(ctx, req.ActingUserId, req.TenantId); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	app, err := s.resolveApp(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.licenses.Revoke(ctx, req.TenantId, app.ID); err != nil {
		return nil, fmt.Errorf("revoking application: %w", err)
	}

	events.TenantAppRevoked.Emit(ctx, events.AppEntitlementEvent{TenantID: req.TenantId, ApplicationID: app.ID})
	return &entitlementpb.RevokeApplicationResponse{}, nil
}

// ListTenantApplications lists applications authorized for a tenant.
func (s *EntitlementServer) ListTenantApplications(ctx context.Context, req *entitlementpb.ListTenantApplicationsRequest) (*entitlementpb.ListTenantApplicationsResponse, error) {
	tas, err := s.licenses.ListByTenant(ctx, req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("listing tenant applications: %w", err)
	}

	result := make([]*entitlementpb.TenantApplication, len(tas))
	for i, ta := range tas {
		result[i] = &entitlementpb.TenantApplication{
			ApplicationId: ta.ApplicationID,
			AuthorizedAt:  ta.CreatedAt.Unix(),
		}
	}

	return &entitlementpb.ListTenantApplicationsResponse{Applications: result}, nil
}

// GrantUserAccess grants a user access to the calling application with roles and scopes.
func (s *EntitlementServer) GrantUserAccess(ctx context.Context, req *entitlementpb.GrantUserAccessRequest) (*entitlementpb.GrantUserAccessResponse, error) {
	if err := s.requireAdmin(ctx, req.ActingUserId, req.TenantId); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	app, err := s.resolveApp(ctx)
	if err != nil {
		return nil, err
	}

	ua, err := s.grants.Grant(ctx, req.UserId, req.TenantId, app.ID, req.Roles, req.Scopes)
	if err != nil {
		return nil, fmt.Errorf("granting access: %w", err)
	}

	_ = ua // response is empty on success
	events.UserAppGranted.Emit(ctx, events.AppEntitlementEvent{TenantID: req.TenantId, ApplicationID: app.ID, UserID: req.UserId})
	return &entitlementpb.GrantUserAccessResponse{}, nil
}

// RevokeUserAccess revokes a user's access to the calling application.
func (s *EntitlementServer) RevokeUserAccess(ctx context.Context, req *entitlementpb.RevokeUserAccessRequest) (*entitlementpb.RevokeUserAccessResponse, error) {
	if err := s.requireAdmin(ctx, req.ActingUserId, req.TenantId); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	app, err := s.resolveApp(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.grants.Revoke(ctx, req.UserId, req.TenantId, app.ID); err != nil {
		return nil, fmt.Errorf("revoking access: %w", err)
	}

	events.UserAppRevoked.Emit(ctx, events.AppEntitlementEvent{TenantID: req.TenantId, ApplicationID: app.ID, UserID: req.UserId})
	return &entitlementpb.RevokeUserAccessResponse{}, nil
}

// UpdateUserAccess updates a user's roles and scopes for the calling application.
func (s *EntitlementServer) UpdateUserAccess(ctx context.Context, req *entitlementpb.UpdateUserAccessRequest) (*entitlementpb.UpdateUserAccessResponse, error) {
	if err := s.requireAdmin(ctx, req.ActingUserId, req.TenantId); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	app, err := s.resolveApp(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := s.grants.UpdateAccess(ctx, req.UserId, req.TenantId, app.ID, req.Roles, req.Scopes); err != nil {
		return nil, fmt.Errorf("updating access: %w", err)
	}

	return &entitlementpb.UpdateUserAccessResponse{}, nil
}

// ListUserAccess lists users with access to the calling application in a tenant.
func (s *EntitlementServer) ListUserAccess(ctx context.Context, req *entitlementpb.ListUserAccessRequest) (*entitlementpb.ListUserAccessResponse, error) {
	app, err := s.resolveApp(ctx)
	if err != nil {
		return nil, err
	}

	uas, err := s.grants.ListByTenantAndApp(ctx, req.TenantId, app.ID)
	if err != nil {
		return nil, fmt.Errorf("listing user access: %w", err)
	}

	result := make([]*entitlementpb.UserAccess, len(uas))
	for i, ua := range uas {
		result[i] = &entitlementpb.UserAccess{
			UserId:    ua.UserID,
			Roles:     ua.Roles,
			Scopes:    ua.Scopes,
			GrantedAt: ua.CreatedAt.Unix(),
		}
	}

	return &entitlementpb.ListUserAccessResponse{Users: result}, nil
}
