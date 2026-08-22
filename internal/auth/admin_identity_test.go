package auth

import (
	"reflect"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

func TestAdminIdentityUnionsScopesAndRoles(t *testing.T) {
	id := NewAdminIdentity("u-1", "op@janus.example", []models.AuthorizedTenant{
		{TenantID: "t-1", AppRoles: []string{"operator"}, AppScopes: []string{"directory:read", "users:manage"}},
		{TenantID: "t-2", AppRoles: []string{"auditor"}, AppScopes: []string{"directory:read", "tenants:manage"}},
	})

	if id.ID() != "u-1" || id.Email() != "op@janus.example" || id.TenantID() != "" {
		t.Fatalf("identity basics wrong: %+v", id)
	}

	// Scopes are the deduped, sorted union.
	want := []string{"directory:read", "tenants:manage", "users:manage"}
	if got := id.Scopes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes = %v, want %v", got, want)
	}
	for _, s := range want {
		if !id.HasScope(s) {
			t.Fatalf("HasScope(%q) = false", s)
		}
	}
	if id.HasScope("applications:manage") {
		t.Fatal("HasScope should be false for an unheld scope")
	}

	// Roles unioned across tenants.
	if !id.HasRole("operator") || !id.HasRole("auditor") || id.HasRole("superadmin") {
		t.Fatalf("roles = %v", id.Roles())
	}
}

func TestAdminIdentityEmpty(t *testing.T) {
	id := NewAdminIdentity("u-1", "x@y.z", nil)
	if len(id.Scopes()) != 0 || len(id.Roles()) != 0 || id.HasScope("anything") {
		t.Fatal("empty entitlements should yield no scopes/roles")
	}
}
