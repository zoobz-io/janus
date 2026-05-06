//go:build testing

package integration

import (
	"context"
	"testing"

	identitypb "github.com/zoobz-io/aegis/proto/identity/v1"
	sessionpb "github.com/zoobz-io/aegis/proto/session/v1"
	directorypb "github.com/zoobz-io/aegis/proto/directory/v1"

	"github.com/zoobz-io/janus/internal/mesh"
)

func TestIdentityServer(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	srv := mesh.NewIdentityServer(
		testStores.Users, testStores.LinkedIdentities, testStores.Tenants, testStores.Memberships,
		testStores.Applications, testStores.TenantApplications, testStores.UserApplications,
	)

	var registeredUserID string

	t.Run("Register", func(t *testing.T) {
		resp, err := srv.Register(ctx, &identitypb.RegisterRequest{
			Email:          "mesh-user@example.com",
			Name:           "Mesh User",
			Provider:       "github",
			ProviderUserId: "gh-mesh-1",
		})
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
		if resp.UserId == "" {
			t.Fatal("expected non-empty user ID")
		}
		if resp.TenantId != "" {
			t.Fatal("expected empty tenant ID when no tenant requested")
		}
		registeredUserID = resp.UserId
	})

	t.Run("RegisterWithTenant", func(t *testing.T) {
		resp, err := srv.Register(ctx, &identitypb.RegisterRequest{
			Email:          "tenant-user@example.com",
			Name:           "Tenant User",
			Provider:       "google",
			ProviderUserId: "goog-tenant-1",
			TenantName:     "NewCorp",
			TenantSlug:     "newcorp",
		})
		if err != nil {
			t.Fatalf("Register with tenant: %v", err)
		}
		if resp.TenantId == "" {
			t.Fatal("expected non-empty tenant ID")
		}

		// Verify owner membership was created.
		mem, _ := testStores.Memberships.GetByUserAndTenant(ctx, resp.UserId, resp.TenantId)
		if mem == nil {
			t.Fatal("expected owner membership")
		}
		if mem.Role != "owner" {
			t.Fatalf("expected role owner, got %s", mem.Role)
		}
	})

	t.Run("ResolveIdentity", func(t *testing.T) {
		resp, err := srv.ResolveIdentity(ctx, &identitypb.ResolveIdentityRequest{
			Provider:       "github",
			ProviderUserId: "gh-mesh-1",
		})
		if err != nil {
			t.Fatalf("ResolveIdentity: %v", err)
		}
		if resp.UserId != registeredUserID {
			t.Fatalf("expected user ID %s, got %s", registeredUserID, resp.UserId)
		}
	})

	t.Run("ResolveIdentityNotFound", func(t *testing.T) {
		_, err := srv.ResolveIdentity(ctx, &identitypb.ResolveIdentityRequest{
			Provider:       "github",
			ProviderUserId: "nonexistent",
		})
		if err == nil {
			t.Fatal("expected error for unknown identity")
		}
	})

	t.Run("ListProviders", func(t *testing.T) {
		resp, err := srv.ListProviders(ctx, &identitypb.ListProvidersRequest{
			UserId: registeredUserID,
		})
		if err != nil {
			t.Fatalf("ListProviders: %v", err)
		}
		if len(resp.Providers) != 1 {
			t.Fatalf("expected 1 provider, got %d", len(resp.Providers))
		}
		if resp.Providers[0].Provider != "github" {
			t.Fatalf("expected provider github, got %s", resp.Providers[0].Provider)
		}
	})
}

func TestSessionServer(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	user, _ := testStores.Users.CreateUser(ctx, "sess-mesh@example.com", "Sess Mesh")
	srv := mesh.NewSessionServer(
		testStores.Sessions, testStores.Users, testStores.Tenants,
		testStores.Applications, testStores.TenantApplications, testStores.UserApplications, testStores.Memberships,
	)

	var validToken string

	t.Run("CreateSession", func(t *testing.T) {
		resp, err := srv.CreateSession(ctx, &sessionpb.CreateSessionRequest{
			UserId:    user.ID,
			UserAgent: "test-agent",
			IpAddress: "10.0.0.1",
		})
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
		if resp.Token == "" {
			t.Fatal("expected non-empty token")
		}
		if resp.ExpiresAt == 0 {
			t.Fatal("expected non-zero expires_at")
		}
		validToken = resp.Token
	})

	t.Run("ValidateSession", func(t *testing.T) {
		resp, err := srv.ValidateSession(ctx, &sessionpb.ValidateSessionRequest{
			Token: validToken,
		})
		if err != nil {
			t.Fatalf("ValidateSession: %v", err)
		}
		if !resp.Valid {
			t.Fatal("expected valid=true")
		}
		if resp.UserId != user.ID {
			t.Fatalf("expected user ID %s, got %s", user.ID, resp.UserId)
		}
	})

	t.Run("ValidateSessionBadToken", func(t *testing.T) {
		resp, err := srv.ValidateSession(ctx, &sessionpb.ValidateSessionRequest{
			Token: "bad-token",
		})
		if err != nil {
			t.Fatalf("ValidateSession: %v", err)
		}
		if resp.Valid {
			t.Fatal("expected valid=false for bad token")
		}
	})

	t.Run("RevokeSession", func(t *testing.T) {
		resp, _ := srv.CreateSession(ctx, &sessionpb.CreateSessionRequest{
			UserId: user.ID,
		})
		_, err := srv.RevokeSession(ctx, &sessionpb.RevokeSessionRequest{
			Token: resp.Token,
		})
		if err != nil {
			t.Fatalf("RevokeSession: %v", err)
		}

		validate, _ := srv.ValidateSession(ctx, &sessionpb.ValidateSessionRequest{
			Token: resp.Token,
		})
		if validate.Valid {
			t.Fatal("expected valid=false after revocation")
		}
	})

	t.Run("RevokeUserSessions", func(t *testing.T) {
		srv.CreateSession(ctx, &sessionpb.CreateSessionRequest{UserId: user.ID})
		srv.CreateSession(ctx, &sessionpb.CreateSessionRequest{UserId: user.ID})

		resp, err := srv.RevokeUserSessions(ctx, &sessionpb.RevokeUserSessionsRequest{
			UserId: user.ID,
		})
		if err != nil {
			t.Fatalf("RevokeUserSessions: %v", err)
		}
		if resp.RevokedCount == 0 {
			t.Fatal("expected at least one revoked")
		}
	})

	t.Run("ListUserSessions", func(t *testing.T) {
		srv.CreateSession(ctx, &sessionpb.CreateSessionRequest{
			UserId:    user.ID,
			UserAgent: "list-agent",
		})

		resp, err := srv.ListUserSessions(ctx, &sessionpb.ListUserSessionsRequest{
			UserId: user.ID,
		})
		if err != nil {
			t.Fatalf("ListUserSessions: %v", err)
		}
		if len(resp.Sessions) == 0 {
			t.Fatal("expected at least one session")
		}
	})
}

func TestDirectoryServer(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	srv := mesh.NewDirectoryServer(testStores.Users, testStores.Tenants, testStores.Memberships)

	user, _ := testStores.Users.CreateUser(ctx, "dir-user@example.com", "Dir User")

	t.Run("GetUser", func(t *testing.T) {
		resp, err := srv.GetUser(ctx, &directorypb.GetUserRequest{UserId: user.ID})
		if err != nil {
			t.Fatalf("GetUser: %v", err)
		}
		if resp.Email != "dir-user@example.com" {
			t.Fatalf("expected dir-user@example.com, got %s", resp.Email)
		}
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		resp, err := srv.GetUserByEmail(ctx, &directorypb.GetUserByEmailRequest{
			Email: "dir-user@example.com",
		})
		if err != nil {
			t.Fatalf("GetUserByEmail: %v", err)
		}
		if resp.Id != user.ID {
			t.Fatalf("expected user ID %s, got %s", user.ID, resp.Id)
		}
	})

	t.Run("CreateTenantWithOwner", func(t *testing.T) {
		resp, err := srv.CreateTenant(ctx, &directorypb.CreateTenantRequest{
			Name:        "DirCorp",
			Slug:        "dircorp",
			OwnerUserId: user.ID,
		})
		if err != nil {
			t.Fatalf("CreateTenant: %v", err)
		}
		if resp.TenantId == "" {
			t.Fatal("expected non-empty tenant ID")
		}

		mem, _ := testStores.Memberships.GetByUserAndTenant(ctx, user.ID, resp.TenantId)
		if mem == nil {
			t.Fatal("expected owner membership")
		}
		if mem.Role != "owner" {
			t.Fatalf("expected role owner, got %s", mem.Role)
		}
	})

	t.Run("GetTenant", func(t *testing.T) {
		tenant, _ := testStores.Tenants.CreateTenant(ctx, "LookupCorp", "lookupcorp")
		resp, err := srv.GetTenant(ctx, &directorypb.GetTenantRequest{TenantId: tenant.ID})
		if err != nil {
			t.Fatalf("GetTenant: %v", err)
		}
		if resp.Slug != "lookupcorp" {
			t.Fatalf("expected slug lookupcorp, got %s", resp.Slug)
		}
	})

	t.Run("UpdateTenant", func(t *testing.T) {
		tenant, _ := testStores.Tenants.CreateTenant(ctx, "OldName", "oldname")
		_, err := srv.UpdateTenant(ctx, &directorypb.UpdateTenantRequest{
			TenantId: tenant.ID,
			Name:     "NewName",
		})
		if err != nil {
			t.Fatalf("UpdateTenant: %v", err)
		}

		got, _ := testStores.Tenants.GetTenant(ctx, tenant.ID)
		if got.Name != "NewName" {
			t.Fatalf("expected NewName, got %s", got.Name)
		}
	})
}
