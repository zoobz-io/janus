//go:build testing

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/internal/authz"
	"github.com/zoobz-io/janus/models"
)

func TestMembershipManagement(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	owner, _ := testStores.Users.CreateUser(ctx, "owner@example.com", "Owner")
	member, _ := testStores.Users.CreateUser(ctx, "member@example.com", "Member")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "MgmtCorp", "mgmtcorp")
	testStores.Memberships.Create(ctx, owner.ID, tenant.ID, models.UserRoleOwner)

	t.Run("AddMember", func(t *testing.T) {
		mem, err := testStores.Memberships.Create(ctx, member.ID, tenant.ID, models.UserRoleViewer)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if mem.Role != models.UserRoleViewer {
			t.Fatalf("expected viewer, got %s", mem.Role)
		}
	})

	t.Run("ListByTenant", func(t *testing.T) {
		result, err := testStores.Memberships.ListByTenant(ctx, tenant.ID, models.OffsetPage{Limit: 100})
		if err != nil {
			t.Fatalf("ListByTenant: %v", err)
		}
		if len(result.Items) != 2 {
			t.Fatalf("expected 2 members, got %d", len(result.Items))
		}
	})

	t.Run("UpdateRole", func(t *testing.T) {
		mem, _ := testStores.Memberships.GetByUserAndTenant(ctx, member.ID, tenant.ID)
		if err := testStores.Memberships.Delete(ctx, mem.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		updated, err := testStores.Memberships.Create(ctx, member.ID, tenant.ID, models.UserRoleAdmin)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if updated.Role != models.UserRoleAdmin {
			t.Fatalf("expected admin, got %s", updated.Role)
		}
	})

	t.Run("RemoveMember", func(t *testing.T) {
		mem, _ := testStores.Memberships.GetByUserAndTenant(ctx, member.ID, tenant.ID)
		if err := testStores.Memberships.Delete(ctx, mem.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		gone, _ := testStores.Memberships.GetByUserAndTenant(ctx, member.ID, tenant.ID)
		if gone != nil {
			t.Fatal("expected nil after removal")
		}
	})
}

func TestEntitlementFlow(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	// Set up: user in a tenant, application exists.
	user, _ := testStores.Users.CreateUser(ctx, "ent-user@example.com", "Ent User")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "EntCorp", "entcorp")
	testStores.Memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleViewer)
	app, _ := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")

	t.Run("NoEntitlementByDefault", func(t *testing.T) {
		// User has no app grants yet.
		uas, err := testStores.UserApplications.ListByUser(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(uas) != 0 {
			t.Fatalf("expected 0 grants, got %d", len(uas))
		}
	})

	t.Run("AuthorizeTenantForApp", func(t *testing.T) {
		ta, err := testStores.TenantApplications.Authorize(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("Authorize: %v", err)
		}
		if ta.ApplicationID != app.ID {
			t.Fatalf("expected app ID %s, got %s", app.ID, ta.ApplicationID)
		}
	})

	t.Run("GrantUserAppAccess", func(t *testing.T) {
		ua, err := testStores.UserApplications.Grant(ctx, user.ID, tenant.ID, app.ID, []string{"editor"}, []string{"projects:write"})
		if err != nil {
			t.Fatalf("Grant: %v", err)
		}
		if ua.UserID != user.ID {
			t.Fatalf("expected user ID %s, got %s", user.ID, ua.UserID)
		}
	})

	t.Run("UserIsEntitled", func(t *testing.T) {
		uas, err := testStores.UserApplications.ListByUser(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(uas) != 1 {
			t.Fatalf("expected 1 grant, got %d", len(uas))
		}
		if uas[0].ApplicationID != app.ID {
			t.Fatalf("expected app %s, got %s", app.ID, uas[0].ApplicationID)
		}
	})

	t.Run("RevokeUserAppAccess", func(t *testing.T) {
		if err := testStores.UserApplications.Revoke(ctx, user.ID, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.UserApplications.GetByUserAndApp(ctx, user.ID, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})

	t.Run("RevokeTenantAppAccess", func(t *testing.T) {
		if err := testStores.TenantApplications.Revoke(ctx, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.TenantApplications.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}

func TestOwnerProtection(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	owner, _ := testStores.Users.CreateUser(ctx, "sole-owner@example.com", "Sole Owner")
	coowner, _ := testStores.Users.CreateUser(ctx, "co-owner@example.com", "Co Owner")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "OwnerCorp", "ownercorp")
	testStores.Memberships.Create(ctx, owner.ID, tenant.ID, models.UserRoleOwner)

	t.Run("CannotRemoveLastOwner", func(t *testing.T) {
		err := authz.RequireOwnerExists(ctx, testStores.Memberships, tenant.ID, owner.ID)
		if err == nil {
			t.Fatal("expected ErrLastOwner")
		}
		if err != authz.ErrLastOwner {
			t.Fatalf("expected ErrLastOwner, got %v", err)
		}
	})

	t.Run("CanRemoveOwnerWhenCoOwnerExists", func(t *testing.T) {
		testStores.Memberships.Create(ctx, coowner.ID, tenant.ID, models.UserRoleOwner)

		err := authz.RequireOwnerExists(ctx, testStores.Memberships, tenant.ID, owner.ID)
		if err != nil {
			t.Fatalf("expected no error with co-owner, got %v", err)
		}
	})

	t.Run("CannotDemoteToLastOwner", func(t *testing.T) {
		// Remove co-owner so only one owner remains.
		mem, _ := testStores.Memberships.GetByUserAndTenant(ctx, coowner.ID, tenant.ID)
		testStores.Memberships.Delete(ctx, mem.ID)

		err := authz.RequireOwnerExists(ctx, testStores.Memberships, tenant.ID, owner.ID)
		if err != authz.ErrLastOwner {
			t.Fatalf("expected ErrLastOwner after co-owner removed, got %v", err)
		}
	})
}
