package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zoobz-io/rocco"

	"github.com/zoobz-io/janus/database/models"
)

type fakeSessions struct {
	sess *models.Session
	err  error
}

func (f fakeSessions) ValidateByToken(_ context.Context, _ string) (*models.Session, error) {
	return f.sess, f.err
}

type fakeUsers struct{ user *models.User }

func (f fakeUsers) GetUser(_ context.Context, _ string) (*models.User, error) {
	return f.user, nil
}

type fakeEntitlements struct {
	err     error
	tenants []models.AuthorizedTenant
}

func (f fakeEntitlements) ForApplication(_ context.Context, _, _ string) (*models.Application, []models.AuthorizedTenant, error) {
	return &models.Application{Slug: "janus-admin"}, f.tenants, f.err
}

func bearerReq(token string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/applications", nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

// grantedTenants is a non-empty entitlement with operator scopes.
func grantedTenants() []models.AuthorizedTenant {
	return []models.AuthorizedTenant{
		{TenantID: "t-1", AppRoles: []string{"operator"}, AppScopes: []string{"users:manage"}},
	}
}

func TestAdminAuthenticatorBearerWrongIssuerRefused(t *testing.T) {
	sessions := fakeSessions{sess: &models.Session{UserID: "u-1", IssuedBy: "nexus"}}
	auth := NewAdminAuthenticator(sessions, fakeUsers{user: &models.User{ID: "u-1"}},
		fakeEntitlements{tenants: grantedTenants()}, nil)

	if _, err := auth(context.Background(), bearerReq("tok")); err == nil {
		t.Fatal("a customer-app-issued bearer must be refused")
	}
}

func TestAdminAuthenticatorBearerAdminIssuerEntitled(t *testing.T) {
	sessions := fakeSessions{sess: &models.Session{UserID: "u-1", IssuedBy: "janus-admin"}}
	auth := NewAdminAuthenticator(sessions, fakeUsers{user: &models.User{ID: "u-1", Email: "op@x"}},
		fakeEntitlements{tenants: grantedTenants()}, nil)

	id, err := auth(context.Background(), bearerReq("tok"))
	if err != nil {
		t.Fatalf("admin-issued + entitled should pass: %v", err)
	}
	if !id.HasScope("users:manage") || id.ID() != "u-1" {
		t.Fatalf("identity missing scopes/id: %+v", id)
	}
}

func TestAdminAuthenticatorZeroEntitlementRejected(t *testing.T) {
	sessions := fakeSessions{sess: &models.Session{UserID: "u-1", IssuedBy: "janus-admin"}}
	auth := NewAdminAuthenticator(sessions, fakeUsers{user: &models.User{ID: "u-1"}},
		fakeEntitlements{tenants: nil}, nil) // no janus-admin grant

	if _, err := auth(context.Background(), bearerReq("tok")); err == nil {
		t.Fatal("a user with no janus-admin entitlement must be rejected at the door")
	}
}

func TestAdminAuthenticatorCookiePath(t *testing.T) {
	cookie := func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return &JanusIdentity{userID: "u-1", email: "op@x"}, nil
	}
	auth := NewAdminAuthenticator(fakeSessions{}, fakeUsers{}, fakeEntitlements{tenants: grantedTenants()}, cookie)

	id, err := auth(context.Background(), httptest.NewRequest(http.MethodGet, "/applications", nil))
	if err != nil {
		t.Fatalf("cookie path should authenticate: %v", err)
	}
	if id.ID() != "u-1" || !id.HasScope("users:manage") {
		t.Fatalf("cookie identity wrong: %+v", id)
	}
}

func TestAdminAuthenticatorNoCredential(t *testing.T) {
	auth := NewAdminAuthenticator(fakeSessions{}, fakeUsers{}, fakeEntitlements{}, nil)
	if _, err := auth(context.Background(), httptest.NewRequest(http.MethodGet, "/x", nil)); err == nil {
		t.Fatal("no cookie and no bearer must fail")
	}
}

func TestAdminAuthenticatorEntitlementErrorPropagates(t *testing.T) {
	sessions := fakeSessions{sess: &models.Session{UserID: "u-1", IssuedBy: "janus-admin"}}
	auth := NewAdminAuthenticator(sessions, fakeUsers{user: &models.User{ID: "u-1"}},
		fakeEntitlements{err: context.DeadlineExceeded}, nil)

	if _, err := auth(context.Background(), bearerReq("tok")); err == nil {
		t.Fatal("entitlement resolution error must not authenticate")
	}
}
