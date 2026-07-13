//go:build testing

package integration

import (
	"context"
	"strings"
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

func TestAccounts(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, _ := testStores.Users.CreateUser(ctx, "linked@example.com", "Linked")

	t.Run("LinkAndResolve", func(t *testing.T) {
		link, err := testStores.Accounts.Link(ctx, user.ID, "github", "gh-123")
		if err != nil {
			t.Fatalf("Link: %v", err)
		}
		if link.Provider != "github" {
			t.Fatalf("expected provider github, got %s", link.Provider)
		}

		found, err := testStores.Accounts.GetByProviderSubject(ctx, "github", "gh-123")
		if err != nil {
			t.Fatalf("GetByProviderSubject: %v", err)
		}
		if found == nil || found.UserID != user.ID {
			t.Fatal("expected to find account for user")
		}
	})

	t.Run("ListByUser", func(t *testing.T) {
		testStores.Accounts.Link(ctx, user.ID, "google", "goog-456")

		identities, err := testStores.Accounts.ListByUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(identities) < 2 {
			t.Fatalf("expected at least 2 identities, got %d", len(identities))
		}
	})

	t.Run("Unlink", func(t *testing.T) {
		found, _ := testStores.Accounts.GetByProviderSubject(ctx, "github", "gh-123")
		if err := testStores.Accounts.Unlink(ctx, found.ID, user.ID); err != nil {
			t.Fatalf("Unlink: %v", err)
		}

		gone, _ := testStores.Accounts.GetByProviderSubject(ctx, "github", "gh-123")
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

func TestLicenses(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	tenant, _ := testStores.Tenants.CreateTenant(ctx, "AppCorp", "appcorp")
	app, _ := testStores.Applications.CreateApplication(ctx, "Nexus", "nexus")

	t.Run("AuthorizeAndList", func(t *testing.T) {
		ta, err := testStores.Licenses.Authorize(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("Authorize: %v", err)
		}
		if ta.TenantID != tenant.ID {
			t.Fatalf("expected tenant ID %s, got %s", tenant.ID, ta.TenantID)
		}

		list, err := testStores.Licenses.ListByTenant(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("ListByTenant: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 authorization, got %d", len(list))
		}
	})

	t.Run("GetByTenantAndApp", func(t *testing.T) {
		found, err := testStores.Licenses.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("GetByTenantAndApp: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find license")
		}

		notFound, err := testStores.Licenses.GetByTenantAndApp(ctx, tenant.ID, "nonexistent")
		if err != nil {
			t.Fatalf("GetByTenantAndApp (not found): %v", err)
		}
		if notFound != nil {
			t.Fatal("expected nil for nonexistent app")
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		if err := testStores.Licenses.Revoke(ctx, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.Licenses.GetByTenantAndApp(ctx, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}

func TestGrants(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, _ := testStores.Users.CreateUser(ctx, "appuser@example.com", "App User")
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "UserAppCorp", "userappcorp")
	app, _ := testStores.Applications.CreateApplication(ctx, "Forge", "forge")
	testStores.Memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleViewer)

	t.Run("GrantAndList", func(t *testing.T) {
		ua, err := testStores.Grants.Grant(ctx, user.ID, tenant.ID, app.ID, []string{"viewer"}, []string{"read"})
		if err != nil {
			t.Fatalf("Grant: %v", err)
		}
		if ua.ApplicationID != app.ID {
			t.Fatalf("expected app ID %s, got %s", app.ID, ua.ApplicationID)
		}

		list, err := testStores.Grants.ListByUser(ctx, user.ID, tenant.ID)
		if err != nil {
			t.Fatalf("ListByUser: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 grant, got %d", len(list))
		}
	})

	t.Run("GetByUserAndApp", func(t *testing.T) {
		found, err := testStores.Grants.GetByUserAndApp(ctx, user.ID, tenant.ID, app.ID)
		if err != nil {
			t.Fatalf("GetByUserAndApp: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find grant")
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		if err := testStores.Grants.Revoke(ctx, user.ID, tenant.ID, app.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		found, _ := testStores.Grants.GetByUserAndApp(ctx, user.ID, tenant.ID, app.ID)
		if found != nil {
			t.Fatal("expected nil after revoke")
		}
	})
}

func TestConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("SeededApertureSchema", func(t *testing.T) {
		got, err := testStores.Config.GetByKey(ctx, "aperture_schema")
		if err != nil {
			t.Fatalf("GetByKey: %v", err)
		}
		if got == nil {
			t.Fatal("expected seeded aperture_schema row, got nil")
		}
		if !strings.Contains(got.Value, "metrics:") {
			t.Fatalf("expected aperture schema yaml, got %q", got.Value)
		}
	})

	t.Run("GetMissingReturnsNil", func(t *testing.T) {
		got, err := testStores.Config.GetByKey(ctx, "does_not_exist")
		if err != nil {
			t.Fatalf("GetByKey: %v", err)
		}
		if got != nil {
			t.Fatalf("expected nil for missing key, got %+v", got)
		}
	})

	t.Run("UpsertInsertThenUpdate", func(t *testing.T) {
		if _, err := testStores.Config.Upsert(ctx, "test_key", "v1"); err != nil {
			t.Fatalf("Upsert insert: %v", err)
		}
		got, err := testStores.Config.GetByKey(ctx, "test_key")
		if err != nil {
			t.Fatalf("GetByKey: %v", err)
		}
		if got == nil || got.Value != "v1" {
			t.Fatalf("expected v1, got %+v", got)
		}

		if _, err := testStores.Config.Upsert(ctx, "test_key", "v2"); err != nil {
			t.Fatalf("Upsert update: %v", err)
		}
		got, err = testStores.Config.GetByKey(ctx, "test_key")
		if err != nil {
			t.Fatalf("GetByKey: %v", err)
		}
		if got == nil || got.Value != "v2" {
			t.Fatalf("expected v2 after update, got %+v", got)
		}
	})
}

func TestTiersScopesFeatures(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	app, _ := testStores.Applications.CreateApplication(ctx, "Forge", "forge")

	read, err := testStores.Scopes.Define(ctx, app.ID, "projects:read", "Read projects")
	if err != nil {
		t.Fatalf("Define read scope: %v", err)
	}
	write, err := testStores.Scopes.Define(ctx, app.ID, "builds:write", "Write builds")
	if err != nil {
		t.Fatalf("Define write scope: %v", err)
	}
	free, _ := testStores.Tiers.Define(ctx, app.ID, "free", "Free", 0)
	pro, _ := testStores.Tiers.Define(ctx, app.ID, "pro", "Pro", 1)

	t.Run("ScopeCatalog", func(t *testing.T) {
		got, err := testStores.Scopes.GetByName(ctx, app.ID, "projects:read")
		if err != nil {
			t.Fatalf("GetByName: %v", err)
		}
		if got == nil || got.ID != read.ID {
			t.Fatalf("expected read scope, got %+v", got)
		}
		list, err := testStores.Scopes.ListByApplication(ctx, app.ID)
		if err != nil {
			t.Fatalf("ListByApplication: %v", err)
		}
		if len(list) != 2 {
			t.Fatalf("expected 2 scopes, got %d", len(list))
		}
	})

	t.Run("TierCatalog", func(t *testing.T) {
		bySlug, err := testStores.Tiers.GetBySlug(ctx, app.ID, "pro")
		if err != nil {
			t.Fatalf("GetBySlug: %v", err)
		}
		if bySlug == nil || bySlug.ID != pro.ID {
			t.Fatalf("expected pro tier, got %+v", bySlug)
		}
		list, err := testStores.Tiers.ListByApplication(ctx, app.ID)
		if err != nil {
			t.Fatalf("ListByApplication: %v", err)
		}
		if len(list) != 2 || list[0].Slug != "free" || list[1].Slug != "pro" {
			t.Fatalf("expected [free, pro] by rank, got %+v", list)
		}
	})

	t.Run("BundleScopesIntoTier", func(t *testing.T) {
		if _, err := testStores.Features.Add(ctx, pro.ID, read.ID); err != nil {
			t.Fatalf("Add read: %v", err)
		}
		if _, err := testStores.Features.Add(ctx, pro.ID, write.ID); err != nil {
			t.Fatalf("Add write: %v", err)
		}
		// Adding a scope already in the tier is idempotent.
		if _, err := testStores.Features.Add(ctx, pro.ID, read.ID); err != nil {
			t.Fatalf("Add read (dup): %v", err)
		}
		feats, err := testStores.Features.ListByTier(ctx, pro.ID)
		if err != nil {
			t.Fatalf("ListByTier: %v", err)
		}
		if len(feats) != 2 {
			t.Fatalf("expected 2 features, got %d", len(feats))
		}

		if err := testStores.Features.Remove(ctx, pro.ID, write.ID); err != nil {
			t.Fatalf("Remove write: %v", err)
		}
		feats, _ = testStores.Features.ListByTier(ctx, pro.ID)
		if len(feats) != 1 || feats[0].ScopeID != read.ID {
			t.Fatalf("expected only read feature left, got %+v", feats)
		}
	})

	t.Run("GrantOnTier", func(t *testing.T) {
		user, _ := testStores.Users.CreateUser(ctx, "tieruser@example.com", "Tier User")
		tenant, _ := testStores.Tenants.CreateTenant(ctx, "TierCorp", "tiercorp")
		testStores.Memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleViewer)
		if _, err := testStores.Grants.Grant(ctx, user.ID, tenant.ID, app.ID, []string{"viewer"}, []string{"projects:read"}); err != nil {
			t.Fatalf("Grant: %v", err)
		}

		g, err := testStores.Grants.SetTier(ctx, user.ID, tenant.ID, app.ID, pro.ID)
		if err != nil {
			t.Fatalf("SetTier: %v", err)
		}
		if g.TierID == nil || *g.TierID != pro.ID {
			t.Fatalf("expected grant on pro tier, got %+v", g.TierID)
		}

		g, err = testStores.Grants.SetTier(ctx, user.ID, tenant.ID, app.ID, "")
		if err != nil {
			t.Fatalf("SetTier clear: %v", err)
		}
		if g.TierID != nil {
			t.Fatalf("expected tier cleared, got %v", *g.TierID)
		}
	})

	t.Run("CascadeOnScopeDelete", func(t *testing.T) {
		// pro currently bundles read. Deleting the scope should cascade its feature.
		if err := testStores.Scopes.Delete(ctx, read.ID); err != nil {
			t.Fatalf("Delete scope: %v", err)
		}
		feats, err := testStores.Features.ListByTier(ctx, pro.ID)
		if err != nil {
			t.Fatalf("ListByTier: %v", err)
		}
		if len(feats) != 0 {
			t.Fatalf("expected features cascaded on scope delete, got %d", len(feats))
		}
	})

	_ = free
}
