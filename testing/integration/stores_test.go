//go:build testing

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/models"
)

func TestUsers(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	t.Run("CreateAndGet", func(t *testing.T) {
		user, err := testStores.Users.CreateUser(ctx, "alice@example.com", "Alice")
		if err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		if user.ID == "" {
			t.Fatal("expected non-empty user ID")
		}
		if user.Email != "alice@example.com" {
			t.Fatalf("expected email alice@example.com, got %s", user.Email)
		}

		got, err := testStores.Users.GetUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUser: %v", err)
		}
		if got.DisplayName != "Alice" {
			t.Fatalf("expected display name Alice, got %s", got.DisplayName)
		}
	})

	t.Run("GetByEmail", func(t *testing.T) {
		got, err := testStores.Users.GetUserByEmail(ctx, "alice@example.com")
		if err != nil {
			t.Fatalf("GetUserByEmail: %v", err)
		}
		if got.Email != "alice@example.com" {
			t.Fatalf("expected alice@example.com, got %s", got.Email)
		}
	})

	t.Run("UpdateDisplayName", func(t *testing.T) {
		user, _ := testStores.Users.GetUserByEmail(ctx, "alice@example.com")
		updated, err := testStores.Users.UpdateDisplayName(ctx, user.ID, "Alice Updated")
		if err != nil {
			t.Fatalf("UpdateDisplayName: %v", err)
		}
		if updated.DisplayName != "Alice Updated" {
			t.Fatalf("expected Alice Updated, got %s", updated.DisplayName)
		}
	})

	t.Run("TouchLastSeen", func(t *testing.T) {
		user, _ := testStores.Users.GetUserByEmail(ctx, "alice@example.com")
		before := user.LastSeenAt
		if err := testStores.Users.TouchLastSeen(ctx, user.ID); err != nil {
			t.Fatalf("TouchLastSeen: %v", err)
		}
		after, _ := testStores.Users.GetUser(ctx, user.ID)
		if after.LastSeenAt == nil {
			t.Fatal("expected last_seen_at to be set")
		}
		if before != nil && !after.LastSeenAt.After(*before) {
			t.Fatal("expected last_seen_at to advance")
		}
	})
}

func TestTenants(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	t.Run("CreateAndGet", func(t *testing.T) {
		tenant, err := testStores.Tenants.CreateTenant(ctx, "Acme", "acme")
		if err != nil {
			t.Fatalf("CreateTenant: %v", err)
		}
		if tenant.Slug != "acme" {
			t.Fatalf("expected slug acme, got %s", tenant.Slug)
		}

		got, err := testStores.Tenants.GetTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenant: %v", err)
		}
		if got.Name != "Acme" {
			t.Fatalf("expected name Acme, got %s", got.Name)
		}
	})

	t.Run("ListTenants", func(t *testing.T) {
		result, err := testStores.Tenants.ListTenants(ctx, models.OffsetPage{Limit: 10})
		if err != nil {
			t.Fatalf("ListTenants: %v", err)
		}
		if len(result.Items) == 0 {
			t.Fatal("expected at least one tenant")
		}
	})
}

func TestSessions(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, err := testStores.Users.CreateUser(ctx, "session-user@example.com", "Session User")
	if err != nil {
		t.Fatalf("setup user: %v", err)
	}

	t.Run("CreateAndValidate", func(t *testing.T) {
		token, sess, err := testStores.Sessions.CreateSession(ctx, user.ID, "janus", "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}
		if sess.UserID != user.ID {
			t.Fatalf("expected user ID %s, got %s", user.ID, sess.UserID)
		}

		valid, err := testStores.Sessions.ValidateByToken(ctx, token)
		if err != nil {
			t.Fatalf("ValidateByToken: %v", err)
		}
		if valid == nil {
			t.Fatal("expected valid session")
		}
		if valid.UserID != user.ID {
			t.Fatalf("expected user ID %s, got %s", user.ID, valid.UserID)
		}
	})

	t.Run("ValidateInvalidToken", func(t *testing.T) {
		valid, err := testStores.Sessions.ValidateByToken(ctx, "bogus-token")
		if err != nil {
			t.Fatalf("ValidateByToken: %v", err)
		}
		if valid != nil {
			t.Fatal("expected nil for invalid token")
		}
	})

	t.Run("RevokeByToken", func(t *testing.T) {
		token, _, err := testStores.Sessions.CreateSession(ctx, user.ID, "janus", "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		if err := testStores.Sessions.RevokeByToken(ctx, token); err != nil {
			t.Fatalf("RevokeByToken: %v", err)
		}

		valid, _ := testStores.Sessions.ValidateByToken(ctx, token)
		if valid != nil {
			t.Fatal("expected nil after revocation")
		}
	})

	t.Run("RevokeUserSessions", func(t *testing.T) {
		testStores.Sessions.CreateSession(ctx, user.ID, "janus", "agent1", "127.0.0.1")
		testStores.Sessions.CreateSession(ctx, user.ID, "janus", "agent2", "127.0.0.1")

		count, err := testStores.Sessions.RevokeUserSessions(ctx, user.ID)
		if err != nil {
			t.Fatalf("RevokeUserSessions: %v", err)
		}
		if count == 0 {
			t.Fatal("expected at least one session revoked")
		}

		sessions, _ := testStores.Sessions.ListSessionsByUser(ctx, user.ID)
		if len(sessions) != 0 {
			t.Fatalf("expected 0 sessions after revoke all, got %d", len(sessions))
		}
	})
}

func TestMemberships(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, _ := testStores.Users.CreateUser(ctx, "member@example.com", "Member")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "MemberCorp", "membercorp")

	t.Run("CreateAndList", func(t *testing.T) {
		mem, err := testStores.Memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleAdmin)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if mem.Role != models.UserRoleAdmin {
			t.Fatalf("expected role admin, got %s", mem.Role)
		}

		mems, err := testStores.Memberships.ListByUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(mems) != 1 {
			t.Fatalf("expected 1 membership, got %d", len(mems))
		}
	})

	t.Run("GetByUserAndTenant", func(t *testing.T) {
		mem, err := testStores.Memberships.GetByUserAndTenant(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Fatalf("GetByUserAndTenant: %v", err)
		}
		if mem == nil {
			t.Fatal("expected membership, got nil")
		}

		noMem, err := testStores.Memberships.GetByUserAndTenant(ctx, user.ID, "nonexistent")
		if err != nil {
			t.Fatalf("GetByUserAndTenant (not found): %v", err)
		}
		if noMem != nil {
			t.Fatal("expected nil for nonexistent tenant")
		}
	})
}

func TestLinkedIdentities(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, _ := testStores.Users.CreateUser(ctx, "linked@example.com", "Linked")

	t.Run("LinkAndResolve", func(t *testing.T) {
		link, err := testStores.LinkedIdentities.Link(ctx, user.ID, "github", "gh-123")
		if err != nil {
			t.Fatalf("Link: %v", err)
		}
		if link.Provider != "github" {
			t.Fatalf("expected provider github, got %s", link.Provider)
		}

		found, err := testStores.LinkedIdentities.GetByProviderSubject(ctx, "github", "gh-123")
		if err != nil {
			t.Fatalf("GetByProviderSubject: %v", err)
		}
		if found == nil || found.UserID != user.ID {
			t.Fatal("expected to find linked identity for user")
		}
	})

	t.Run("ListByUser", func(t *testing.T) {
		testStores.LinkedIdentities.Link(ctx, user.ID, "google", "goog-456")

		identities, err := testStores.LinkedIdentities.ListByUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(identities) < 2 {
			t.Fatalf("expected at least 2 identities, got %d", len(identities))
		}
	})

	t.Run("Unlink", func(t *testing.T) {
		found, _ := testStores.LinkedIdentities.GetByProviderSubject(ctx, "github", "gh-123")
		if err := testStores.LinkedIdentities.Unlink(ctx, found.ID, user.ID); err != nil {
			t.Fatalf("Unlink: %v", err)
		}

		gone, _ := testStores.LinkedIdentities.GetByProviderSubject(ctx, "github", "gh-123")
		if gone != nil {
			t.Fatal("expected nil after unlink")
		}
	})
}

func TestApplications(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	t.Run("CreateAndGet", func(t *testing.T) {
		app, err := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")
		if err != nil {
			t.Fatalf("CreateApplication: %v", err)
		}
		if app.Slug != "nexus" {
			t.Fatalf("expected slug nexus, got %s", app.Slug)
		}

		got, err := testStores.Applications.GetApplication(ctx, app.ID)
		if err != nil {
			t.Fatalf("GetApplication: %v", err)
		}
		if got.Name != "Nexus" {
			t.Fatalf("expected name Nexus, got %s", got.Name)
		}
	})

	t.Run("GetBySlug", func(t *testing.T) {
		got, err := testStores.Applications.GetBySlug(ctx, "nexus")
		if err != nil {
			t.Fatalf("GetBySlug: %v", err)
		}
		if got.Name != "Nexus" {
			t.Fatalf("expected name Nexus, got %s", got.Name)
		}
	})

	t.Run("ListApplications", func(t *testing.T) {
		apps, err := testStores.Applications.ListApplications(ctx)
		if err != nil {
			t.Fatalf("ListApplications: %v", err)
		}
		if len(apps) == 0 {
			t.Fatal("expected at least one application")
		}
	})
}

func TestTenantApplications(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	tenant, _ := testStores.Tenants.CreateTenant(ctx, "AppCorp", "appcorp")
	app, _ := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")

	t.Run("AuthorizeAndList", func(t *testing.T) {
		ta, err := testStores.TenantApplications.Authorize(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("Authorize: %v", err)
		}
		if ta.TenantID != tenant.ID {
			t.Fatalf("expected tenant ID %s, got %s", tenant.ID, ta.TenantID)
		}

		list, err := testStores.TenantApplications.ListByTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("ListByTenant: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 authorization, got %d", len(list))
		}
	})

	t.Run("GetByTenantAndApp", func(t *testing.T) {
		found, err := testStores.TenantApplications.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("GetByTenantAndApp: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find tenant application")
		}

		notFound, err := testStores.TenantApplications.GetByTenantAndApp(ctx, tenant.ID, "nonexistent")
		if err != nil {
			t.Fatalf("GetByTenantAndApp (not found): %v", err)
		}
		if notFound != nil {
			t.Fatal("expected nil for nonexistent app")
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		if err := testStores.TenantApplications.Revoke(ctx, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.TenantApplications.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}

func TestUserApplications(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, _ := testStores.Users.CreateUser(ctx, "appuser@example.com", "App User")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "UserAppCorp", "userappcorp")
	app, _ := testStores.Applications.CreateApplication(ctx, "Forge", "forge")
	testStores.Memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleViewer)

	t.Run("GrantAndList", func(t *testing.T) {
		ua, err := testStores.UserApplications.Grant(ctx, user.ID, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("Grant: %v", err)
		}
		if ua.ApplicationID != app.ID {
			t.Fatalf("expected app ID %s, got %s", app.ID, ua.ApplicationID)
		}

		list, err := testStores.UserApplications.ListByUser(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 grant, got %d", len(list))
		}
	})

	t.Run("GetByUserAndApp", func(t *testing.T) {
		found, err := testStores.UserApplications.GetByUserAndApp(ctx, user.ID, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("GetByUserAndApp: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find user application")
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		if err := testStores.UserApplications.Revoke(ctx, user.ID, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.UserApplications.GetByUserAndApp(ctx, user.ID, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}
