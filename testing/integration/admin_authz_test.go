//go:build testing

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/internal/auth"
	"github.com/zoobz-io/janus/internal/authz"
)

// adminAuthenticator builds the real admin authenticator (bearer-only; nil
// cookie extractor) over the test stores.
func adminAuthenticator() func(context.Context, *http.Request) (interface {
	ID() string
	HasScope(string) bool
}, error) {
	ent := authz.NewEntitlements(
		testStores.Applications, testStores.Licenses, testStores.Grants,
		testStores.Memberships, testStores.Tenants, testStores.Features, testStores.Scopes,
		testStores.Users,
	)
	real := auth.NewAdminAuthenticator(testStores.Sessions, testStores.Users, ent, nil)
	return func(ctx context.Context, r *http.Request) (interface {
		ID() string
		HasScope(string) bool
	}, error) {
		id, err := real(ctx, r)
		if err != nil {
			return nil, err
		}
		return id, nil
	}
}

func bearer(t *testing.T, userID, issuedBy string) *http.Request {
	t.Helper()
	token, _, err := testStores.Sessions.CreateSession(context.Background(), userID, issuedBy, "test", "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	r := httptest.NewRequest(http.MethodGet, "/applications", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}

// seedAdminOperator creates the janus-admin app, an active licensed tenant, an
// active user, membership, and an operator grant. Returns the user id.
func seedAdminOperator(t *testing.T) (userID, tenantID, appID string) {
	t.Helper()
	ctx := context.Background()
	app, err := testStores.Applications.CreateApplication(ctx, "Janus Admin", "janus-admin")
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	tenant, _ := testStores.Tenants.CreateTenant(ctx, "Janus Ops", "janus-ops")
	user, _ := testStores.Users.CreateUser(ctx, "op@janus.example", "Op")
	testStores.Memberships.Create(ctx, user.ID, tenant.ID, models.UserRoleOwner)
	if _, err := testStores.Licenses.Authorize(ctx, tenant.ID, app.ID); err != nil {
		t.Fatalf("license: %v", err)
	}
	if _, err := testStores.Grants.Grant(ctx, user.ID, tenant.ID, app.ID,
		[]string{"operator"}, []string{"directory:read", "users:manage"}); err != nil {
		t.Fatalf("grant: %v", err)
	}
	return user.ID, tenant.ID, app.ID
}

func TestAdminAuthzDoor(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })
	authn := adminAuthenticator()

	userID, tenantID, appID := seedAdminOperator(t)

	t.Run("EntitledOperatorAuthenticates", func(t *testing.T) {
		id, err := authn(ctx, bearer(t, userID, "janus-admin"))
		if err != nil {
			t.Fatalf("entitled operator refused: %v", err)
		}
		if !id.HasScope("users:manage") {
			t.Fatal("operator identity missing granted scope")
		}
	})

	t.Run("MeshAppIssuedBearerRefused", func(t *testing.T) {
		// Same entitled user, but the session was minted by a customer app.
		if _, err := authn(ctx, bearer(t, userID, "nexus")); err == nil {
			t.Fatal("a customer-app-issued session is not an admin credential")
		}
	})

	t.Run("MembershipButNoGrantRejected", func(t *testing.T) {
		u, _ := testStores.Users.CreateUser(ctx, "nogrant@janus.example", "NoGrant")
		testStores.Memberships.Create(ctx, u.ID, tenantID, models.UserRoleViewer)
		// Licensed tenant, member — but no janus-admin grant.
		if _, err := authn(ctx, bearer(t, u.ID, "janus-admin")); err == nil {
			t.Fatal("a member without a grant must be rejected at the door")
		}
	})

	t.Run("UnlicensedTenantRejected", func(t *testing.T) {
		other, _ := testStores.Tenants.CreateTenant(ctx, "Unlicensed", "unlic")
		u, _ := testStores.Users.CreateUser(ctx, "unlic@janus.example", "Unlic")
		testStores.Memberships.Create(ctx, u.ID, other.ID, models.UserRoleOwner)
		// Grant exists but the tenant is not licensed for janus-admin.
		testStores.Grants.Grant(ctx, u.ID, other.ID, appID, []string{"operator"}, []string{"users:manage"})
		if _, err := authn(ctx, bearer(t, u.ID, "janus-admin")); err == nil {
			t.Fatal("a grant in an unlicensed tenant must not entitle admin access")
		}
	})

	t.Run("DeactivatedUserRefused", func(t *testing.T) {
		// Deactivate the operator; their live session must stop working.
		if _, err := testStores.Users.Update(ctx, userID, "Op", models.UserStatusInactive); err != nil {
			t.Fatalf("deactivate: %v", err)
		}
		if _, err := authn(ctx, bearer(t, userID, "janus-admin")); err == nil {
			t.Fatal("a deactivated user with a live session must be refused")
		}
		// Reactivate for the suspended-tenant case.
		testStores.Users.Update(ctx, userID, "Op", models.UserStatusActive)
	})

	t.Run("SuspendedTenantRefused", func(t *testing.T) {
		if _, err := testStores.Tenants.UpdateTenant(ctx, tenantID, "Janus Ops", models.TenantStatusSuspended); err != nil {
			t.Fatalf("suspend: %v", err)
		}
		if _, err := authn(ctx, bearer(t, userID, "janus-admin")); err == nil {
			t.Fatal("a suspended tenant must not entitle its members")
		}
	})
}
