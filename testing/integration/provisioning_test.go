//go:build testing

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

// tenantBySlug looks a tenant up by slug for existence assertions.
func tenantBySlug(t *testing.T, slug string) *models.Tenant {
	t.Helper()
	results, err := testStores.Tenants.Query().
		Where("slug", "=", "slug").
		Limit(1).
		Exec(context.Background(), map[string]any{"slug": slug})
	if err != nil {
		t.Fatalf("querying tenant by slug: %v", err)
	}
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

func TestCreateTenantWithOwner(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	owner, _ := testStores.Users.CreateUser(ctx, "prov-owner@example.com", "Prov Owner")

	t.Run("CreatesTenantAndOwnerMembership", func(t *testing.T) {
		tenant, mem, err := testStores.CreateTenantWithOwner(ctx, "ProvCorp", "provcorp", owner.ID)
		if err != nil {
			t.Fatalf("CreateTenantWithOwner: %v", err)
		}
		if mem.Role != models.UserRoleOwner {
			t.Fatalf("expected owner role, got %s", mem.Role)
		}
		fetched, err := testStores.Memberships.GetByUserAndTenant(ctx, owner.ID, tenant.ID)
		if err != nil || fetched == nil {
			t.Fatalf("owner membership not persisted: %v", err)
		}
	})

	t.Run("RollsBackTenantWhenOwnerInvalid", func(t *testing.T) {
		// A nonexistent owner trips the memberships FK — the second write of
		// the transaction — so the tenant from the first write must vanish.
		_, _, err := testStores.CreateTenantWithOwner(ctx, "GhostCorp", "ghostcorp", "no-such-user")
		if err == nil {
			t.Fatal("expected FK violation for nonexistent owner")
		}
		if tenantBySlug(t, "ghostcorp") != nil {
			t.Fatal("tenant persisted despite failed owner membership — transaction did not roll back")
		}
	})
}

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanAll(t) })

	t.Run("CreatesUserAndLink", func(t *testing.T) {
		user, tenant, err := testStores.RegisterUser(ctx, "reg-1@example.com", "Reg One", "github", "gh-reg-1", "", "")
		if err != nil {
			t.Fatalf("RegisterUser: %v", err)
		}
		if tenant != nil {
			t.Fatal("expected no tenant when none requested")
		}
		account, err := testStores.Accounts.GetByProviderSubject(ctx, "github", "gh-reg-1")
		if err != nil || account == nil || account.UserID != user.ID {
			t.Fatalf("account link not persisted: %v", err)
		}
	})

	t.Run("CreatesOwnedTenantWhenRequested", func(t *testing.T) {
		user, tenant, err := testStores.RegisterUser(ctx, "reg-2@example.com", "Reg Two", "github", "gh-reg-2", "RegCorp", "regcorp")
		if err != nil {
			t.Fatalf("RegisterUser with tenant: %v", err)
		}
		if tenant == nil {
			t.Fatal("expected tenant")
		}
		mem, err := testStores.Memberships.GetByUserAndTenant(ctx, user.ID, tenant.ID)
		if err != nil || mem == nil || mem.Role != models.UserRoleOwner {
			t.Fatalf("owner membership missing or wrong role: %v", err)
		}
	})

	t.Run("RollsBackUserWhenLinkTaken", func(t *testing.T) {
		// The (provider, subject) pair is already linked to reg-1, so the
		// account insert — the second write — trips the unique constraint.
		// The user from the first write must vanish.
		_, _, err := testStores.RegisterUser(ctx, "reg-3@example.com", "Reg Three", "github", "gh-reg-1", "", "")
		if err == nil {
			t.Fatal("expected unique violation for taken provider/subject")
		}
		user, lookupErr := testStores.Users.GetUserByEmail(ctx, "reg-3@example.com")
		if lookupErr == nil && user != nil {
			t.Fatal("user persisted despite failed identity link — transaction did not roll back")
		}
	})

	t.Run("RollsBackUserAndLinkWhenTenantSlugTaken", func(t *testing.T) {
		// regcorp exists from the earlier subtest, so the tenant insert — the
		// third write — trips the unique slug. The user AND the account link
		// from the first two writes must both vanish.
		_, _, err := testStores.RegisterUser(ctx, "reg-4@example.com", "Reg Four", "github", "gh-reg-4", "RegCorp Again", "regcorp")
		if err == nil {
			t.Fatal("expected unique violation for taken tenant slug")
		}
		user, lookupErr := testStores.Users.GetUserByEmail(ctx, "reg-4@example.com")
		if lookupErr == nil && user != nil {
			t.Fatal("user persisted despite failed tenant creation — transaction did not roll back")
		}
		account, acctErr := testStores.Accounts.GetByProviderSubject(ctx, "github", "gh-reg-4")
		if acctErr == nil && account != nil {
			t.Fatal("account link persisted despite failed tenant creation — transaction did not roll back")
		}
	})
}
