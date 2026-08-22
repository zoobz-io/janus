package authz

import (
	"context"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

// Stub lookups for driving ForApplication without a database.
type stubApps struct{ app *models.Application }

func (s stubApps) GetBySlug(_ context.Context, _ string) (*models.Application, error) {
	return s.app, nil
}

type stubLicenses struct{ licensed map[string]bool } // keyed by tenantID
func (s stubLicenses) GetByTenantAndApp(_ context.Context, tenantID, _ string) (*models.License, error) {
	if s.licensed[tenantID] {
		return &models.License{TenantID: tenantID}, nil
	}
	return nil, nil
}

type stubGrants struct{ granted map[string]*models.Grant } // keyed by tenantID
func (s stubGrants) GetByUserAndApp(_ context.Context, _, tenantID, _ string) (*models.Grant, error) {
	return s.granted[tenantID], nil
}

type stubMems struct{ mems []*models.Membership }

func (s stubMems) ListByUser(_ context.Context, _ string) ([]*models.Membership, error) {
	return s.mems, nil
}

type stubTenants struct{ tenants map[string]*models.Tenant }

func (s stubTenants) GetTenant(_ context.Context, id string) (*models.Tenant, error) {
	return s.tenants[id], nil
}

type stubFeatures struct{}

func (stubFeatures) ListByTier(_ context.Context, _ string) ([]*models.Feature, error) {
	return nil, nil
}

type stubScopes struct{}

func (stubScopes) GetByID(_ context.Context, _ string) (*models.Scope, error) { return nil, nil }

type stubUsers struct{ user *models.User }

func (s stubUsers) GetUser(_ context.Context, _ string) (*models.User, error) { return s.user, nil }

// entitlementsFor builds a resolver where the user is a member of every tenant
// in tenantStatus, licensed+granted for the app, with the given user status.
func entitlementsFor(userStatus string, tenantStatus map[string]string) *Entitlements {
	app := &models.Application{ID: "app-1", Slug: "janus-admin"}
	mems := make([]*models.Membership, 0, len(tenantStatus))
	tenants := map[string]*models.Tenant{}
	licensed := map[string]bool{}
	granted := map[string]*models.Grant{}
	for id, status := range tenantStatus {
		mems = append(mems, &models.Membership{TenantID: id, Role: models.UserRoleOwner})
		tenants[id] = &models.Tenant{ID: id, Name: id, Status: status}
		licensed[id] = true
		granted[id] = &models.Grant{TenantID: id, Roles: []string{"operator"}, Scopes: []string{"users:manage"}}
	}
	return NewEntitlements(
		stubApps{app: app},
		stubLicenses{licensed: licensed},
		stubGrants{granted: granted},
		stubMems{mems: mems},
		stubTenants{tenants: tenants},
		stubFeatures{},
		stubScopes{},
		stubUsers{user: &models.User{ID: "u-1", Status: userStatus}},
	)
}

func TestForApplicationActiveUser(t *testing.T) {
	e := entitlementsFor(models.UserStatusActive, map[string]string{"t-1": models.TenantStatusActive})
	_, tenants, err := e.ForApplication(context.Background(), "u-1", "janus-admin")
	if err != nil {
		t.Fatalf("ForApplication: %v", err)
	}
	if len(tenants) != 1 || tenants[0].AppScopes[0] != "users:manage" {
		t.Fatalf("active user should resolve entitlements, got %+v", tenants)
	}
}

func TestForApplicationDeactivatedUser(t *testing.T) {
	e := entitlementsFor(models.UserStatusInactive, map[string]string{"t-1": models.TenantStatusActive})
	app, tenants, err := e.ForApplication(context.Background(), "u-1", "janus-admin")
	if err != nil {
		t.Fatalf("ForApplication: %v", err)
	}
	if app == nil {
		t.Fatal("app should still resolve")
	}
	if len(tenants) != 0 {
		t.Fatalf("deactivated user must have no entitlements, got %+v", tenants)
	}
}

func TestForApplicationSuspendedTenantSkipped(t *testing.T) {
	e := entitlementsFor(models.UserStatusActive, map[string]string{
		"t-active":    models.TenantStatusActive,
		"t-suspended": models.TenantStatusSuspended,
	})
	_, tenants, err := e.ForApplication(context.Background(), "u-1", "janus-admin")
	if err != nil {
		t.Fatalf("ForApplication: %v", err)
	}
	if len(tenants) != 1 || tenants[0].TenantID != "t-active" {
		t.Fatalf("suspended tenant must be skipped, got %+v", tenants)
	}
}
