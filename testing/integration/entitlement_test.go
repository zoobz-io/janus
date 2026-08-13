//go:build testing

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/aegis"
	entitlementpb "github.com/zoobz-io/aegis/proto/entitlement/v1"

	"github.com/zoobz-io/janus/internal/authz"
	"github.com/zoobz-io/janus/internal/mesh"
	"github.com/zoobz-io/janus/database/models"
)

// testCtxWithApp returns a context with a security context that identifies
// the caller as the given application slug (cert CN).
func testCtxWithApp(ctx context.Context, appSlug string) context.Context {
	sc := &aegis.SecurityContext{
		Metadata: aegis.Metadata{NodeID: appSlug},
	}
	return aegis.WithTestSecurityContext(ctx, sc)
}

func TestEntitlementAuthz(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	owner, _ := testStores.Users.CreateUser(ctx, "ent-owner@example.com", "Owner")
	viewer, _ := testStores.Users.CreateUser(ctx, "ent-viewer@example.com", "Viewer")
	outsider, _ := testStores.Users.CreateUser(ctx, "ent-outsider@example.com", "Outsider")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "EntAuthzCorp", "entauthzcorp")
	app, _ := testStores.Applications.CreateApplication(ctx, "TestApp", "testapp")

	testStores.Memberships.Create(ctx, owner.ID, tenant.ID, models.UserRoleOwner)
	testStores.Memberships.Create(ctx, viewer.ID, tenant.ID, models.UserRoleViewer)

	t.Run("AdminCanAuthorizeApp", func(t *testing.T) {
		_, err := authz.RequireRole(ctx, testStores.Memberships, owner.ID, tenant.ID, models.UserRoleAdmin, models.UserRoleOwner)
		if err != nil {
			t.Fatalf("owner should pass admin check: %v", err)
		}

		ta, err := testStores.Licenses.Authorize(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("Authorize: %v", err)
		}
		if ta.ApplicationID != app.ID {
			t.Fatalf("expected app %s, got %s", app.ID, ta.ApplicationID)
		}
	})

	t.Run("ViewerCannotAuthorizeApp", func(t *testing.T) {
		_, err := authz.RequireRole(ctx, testStores.Memberships, viewer.ID, tenant.ID, models.UserRoleAdmin, models.UserRoleOwner)
		if err != authz.ErrInsufficientRole {
			t.Fatalf("expected ErrInsufficientRole, got %v", err)
		}
	})

	t.Run("OutsiderCannotAuthorizeApp", func(t *testing.T) {
		_, err := authz.RequireRole(ctx, testStores.Memberships, outsider.ID, tenant.ID, models.UserRoleAdmin, models.UserRoleOwner)
		if err != authz.ErrNotMember {
			t.Fatalf("expected ErrNotMember, got %v", err)
		}
	})

	t.Run("GrantWithRolesAndScopes", func(t *testing.T) {
		roles := []string{"editor", "reviewer"}
		scopes := []string{"projects:read", "builds:write"}

		ua, err := testStores.Grants.Grant(ctx, viewer.ID, tenant.ID, app.ID, roles, scopes)
		if err != nil {
			t.Fatalf("Grant: %v", err)
		}
		if len(ua.Roles) != 2 || ua.Roles[0] != "editor" {
			t.Fatalf("expected roles [editor reviewer], got %v", ua.Roles)
		}
		if len(ua.Scopes) != 2 || ua.Scopes[1] != "builds:write" {
			t.Fatalf("expected scopes [projects:read builds:write], got %v", ua.Scopes)
		}
	})

	t.Run("GetByUserAndAppReturnsRolesScopes", func(t *testing.T) {
		ua, err := testStores.Grants.GetByUserAndApp(ctx, viewer.ID, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("GetByUserAndApp: %v", err)
		}
		if ua == nil {
			t.Fatal("expected grant, got nil")
		}
		if len(ua.Roles) != 2 {
			t.Fatalf("expected 2 roles, got %d", len(ua.Roles))
		}
		if len(ua.Scopes) != 2 {
			t.Fatalf("expected 2 scopes, got %d", len(ua.Scopes))
		}
	})

	t.Run("UpdateAccess", func(t *testing.T) {
		updated, err := testStores.Grants.UpdateAccess(ctx, viewer.ID, tenant.ID, app.ID, []string{"admin"}, []string{"*"})
		if err != nil {
			t.Fatalf("UpdateAccess: %v", err)
		}
		if len(updated.Roles) != 1 || updated.Roles[0] != "admin" {
			t.Fatalf("expected roles [admin], got %v", updated.Roles)
		}
		if len(updated.Scopes) != 1 || updated.Scopes[0] != "*" {
			t.Fatalf("expected scopes [*], got %v", updated.Scopes)
		}

		// Verify persistence.
		fetched, _ := testStores.Grants.GetByUserAndApp(ctx, viewer.ID, tenant.ID, app.ID)
		if len(fetched.Roles) != 1 || fetched.Roles[0] != "admin" {
			t.Fatalf("expected persisted roles [admin], got %v", fetched.Roles)
		}
	})

	t.Run("ListByTenantAndApp", func(t *testing.T) {
		uas, err := testStores.Grants.ListByTenantAndApp(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("ListByTenantAndApp: %v", err)
		}
		if len(uas) != 1 {
			t.Fatalf("expected 1 user with access, got %d", len(uas))
		}
		if uas[0].UserID != viewer.ID {
			t.Fatalf("expected user %s, got %s", viewer.ID, uas[0].UserID)
		}
	})

	t.Run("GrantEmptyRolesScopes", func(t *testing.T) {
		testStores.Grants.Revoke(ctx, viewer.ID, tenant.ID, app.ID)
		ua, err := testStores.Grants.Grant(ctx, viewer.ID, tenant.ID, app.ID, nil, nil)
		if err != nil {
			t.Fatalf("Grant with nil: %v", err)
		}
		if ua.Roles != nil && len(ua.Roles) != 0 {
			t.Fatalf("expected empty roles, got %v", ua.Roles)
		}
	})

	t.Run("RevokeApp", func(t *testing.T) {
		if err := testStores.Licenses.Revoke(ctx, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.Licenses.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}

func TestEntitlementServer(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	owner, _ := testStores.Users.CreateUser(ctx, "esrv-owner@example.com", "Owner")
	viewer, _ := testStores.Users.CreateUser(ctx, "esrv-viewer@example.com", "Viewer")
	target, _ := testStores.Users.CreateUser(ctx, "esrv-target@example.com", "Target")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "ESrvCorp", "esrvcorp")
	app, _ := testStores.Applications.CreateApplication(ctx, "ESrvApp", "esrvapp")

	testStores.Memberships.Create(ctx, owner.ID, tenant.ID, models.UserRoleOwner)
	testStores.Memberships.Create(ctx, viewer.ID, tenant.ID, models.UserRoleViewer)
	testStores.Memberships.Create(ctx, target.ID, tenant.ID, models.UserRoleViewer)

	srv := mesh.NewEntitlementServer(
		testStores.Applications, testStores.Licenses, testStores.Grants, testStores.Memberships,
	)

	// Inject security context so resolveApp finds the application by slug.
	appCtx := testCtxWithApp(ctx, "esrvapp")

	t.Run("AuthorizeApplication", func(t *testing.T) {
		_, err := srv.AuthorizeApplication(appCtx, &entitlementpb.AuthorizeApplicationRequest{
			TenantId:     tenant.ID,
			ActingUserId: owner.ID,
		})
		if err != nil {
			t.Fatalf("AuthorizeApplication: %v", err)
		}

		tas, _ := testStores.Licenses.ListByTenant(ctx, tenant.ID)
		if len(tas) != 1 {
			t.Fatalf("expected 1 tenant app, got %d", len(tas))
		}
	})

	t.Run("AuthorizeApplicationDeniedForViewer", func(t *testing.T) {
		_, err := srv.AuthorizeApplication(appCtx, &entitlementpb.AuthorizeApplicationRequest{
			TenantId:     tenant.ID,
			ActingUserId: viewer.ID,
		})
		if err == nil {
			t.Fatal("expected error for viewer")
		}
	})

	t.Run("GrantUserAccess", func(t *testing.T) {
		_, err := srv.GrantUserAccess(appCtx, &entitlementpb.GrantUserAccessRequest{
			TenantId:     tenant.ID,
			UserId:       target.ID,
			Roles:        []string{"editor"},
			Scopes:       []string{"projects:read", "builds:write"},
			ActingUserId: owner.ID,
		})
		if err != nil {
			t.Fatalf("GrantUserAccess: %v", err)
		}
	})

	t.Run("GrantUserAccessDeniedForViewer", func(t *testing.T) {
		_, err := srv.GrantUserAccess(appCtx, &entitlementpb.GrantUserAccessRequest{
			TenantId:     tenant.ID,
			UserId:       target.ID,
			Roles:        []string{"admin"},
			Scopes:       []string{"*"},
			ActingUserId: viewer.ID,
		})
		if err == nil {
			t.Fatal("expected error for viewer")
		}
	})

	t.Run("ListUserAccess", func(t *testing.T) {
		resp, err := srv.ListUserAccess(appCtx, &entitlementpb.ListUserAccessRequest{
			TenantId: tenant.ID,
		})
		if err != nil {
			t.Fatalf("ListUserAccess: %v", err)
		}
		if len(resp.Users) != 1 {
			t.Fatalf("expected 1 user, got %d", len(resp.Users))
		}
		if resp.Users[0].UserId != target.ID {
			t.Fatalf("expected user %s, got %s", target.ID, resp.Users[0].UserId)
		}
		if len(resp.Users[0].Roles) != 1 || resp.Users[0].Roles[0] != "editor" {
			t.Fatalf("expected roles [editor], got %v", resp.Users[0].Roles)
		}
	})

	t.Run("UpdateUserAccess", func(t *testing.T) {
		_, err := srv.UpdateUserAccess(appCtx, &entitlementpb.UpdateUserAccessRequest{
			TenantId:     tenant.ID,
			UserId:       target.ID,
			Roles:        []string{"admin"},
			Scopes:       []string{"*"},
			ActingUserId: owner.ID,
		})
		if err != nil {
			t.Fatalf("UpdateUserAccess: %v", err)
		}

		resp, _ := srv.ListUserAccess(appCtx, &entitlementpb.ListUserAccessRequest{TenantId: tenant.ID})
		if resp.Users[0].Roles[0] != "admin" {
			t.Fatalf("expected role admin after update, got %v", resp.Users[0].Roles)
		}
	})

	t.Run("RevokeUserAccess", func(t *testing.T) {
		_, err := srv.RevokeUserAccess(appCtx, &entitlementpb.RevokeUserAccessRequest{
			TenantId:     tenant.ID,
			UserId:       target.ID,
			ActingUserId: owner.ID,
		})
		if err != nil {
			t.Fatalf("RevokeUserAccess: %v", err)
		}

		resp, _ := srv.ListUserAccess(appCtx, &entitlementpb.ListUserAccessRequest{TenantId: tenant.ID})
		if len(resp.Users) != 0 {
			t.Fatalf("expected 0 users after revoke, got %d", len(resp.Users))
		}
	})

	t.Run("RevokeApplication", func(t *testing.T) {
		_, err := srv.RevokeApplication(appCtx, &entitlementpb.RevokeApplicationRequest{
			TenantId:     tenant.ID,
			ActingUserId: owner.ID,
		})
		if err != nil {
			t.Fatalf("RevokeApplication: %v", err)
		}

		found, _ := testStores.Licenses.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}
