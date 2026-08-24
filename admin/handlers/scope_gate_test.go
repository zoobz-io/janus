package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/admin/contracts"
	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// fakeUsers satisfies contracts.Users so authorized requests reach a working
// handler. It returns benign data; the tests assert on the scope gate, not on
// store behavior.
type fakeUsers struct{}

func (fakeUsers) GetUser(_ context.Context, _ string) (*models.User, error) {
	return &models.User{ID: "u-1", Email: "op@x", Status: models.UserStatusActive}, nil
}
func (fakeUsers) GetUserByEmail(_ context.Context, _ string) (*models.User, error) {
	return &models.User{ID: "u-1", Email: "op@x"}, nil
}
func (fakeUsers) List(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.User], error) {
	return &models.OffsetResult[models.User]{}, nil
}
func (fakeUsers) Search(_ context.Context, _ stores.UserSearchParams) (*stores.UserSearchResult, error) {
	return &stores.UserSearchResult{}, nil
}
func (fakeUsers) CreateUser(_ context.Context, _, _ string) (*models.User, error) {
	return &models.User{ID: "u-1", Email: "op@x"}, nil
}
func (fakeUsers) Update(_ context.Context, _, _ string, _ models.UserStatus) (*models.User, error) {
	return &models.User{ID: "u-1", Email: "op@x"}, nil
}

func TestMain(m *testing.M) {
	sum.New()
	k := sum.Start()
	sum.Register[contracts.Users](k, fakeUsers{})
	sum.Freeze(k)
	os.Exit(m.Run())
}

// scopedIdentity is a rocco.Identity carrying a fixed scope set.
type scopedIdentity struct{ scopes map[string]bool }

func (scopedIdentity) ID() string               { return "u-1" }
func (scopedIdentity) TenantID() string         { return "" }
func (scopedIdentity) Email() string            { return "op@x" }
func (s scopedIdentity) Scopes() []string       { return nil }
func (scopedIdentity) Roles() []string          { return nil }
func (s scopedIdentity) HasScope(x string) bool { return s.scopes[x] }
func (scopedIdentity) HasRole(string) bool      { return false }
func (scopedIdentity) Stats() map[string]int    { return nil }

func operator() scopedIdentity {
	return scopedIdentity{scopes: map[string]bool{
		"directory:read": true, "users:manage": true,
		"tenants:manage": true, "applications:manage": true,
	}}
}

func auditor() scopedIdentity {
	return scopedIdentity{scopes: map[string]bool{"directory:read": true}}
}

// serve builds an engine authenticating as id and drives one request, returning
// the HTTP status.
func serve(t *testing.T, id rocco.Identity, method, target, body string) int {
	t.Helper()
	eng := rocco.NewEngine().
		WithAuthenticator(func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
			return id, nil
		}).
		WithHandlers(All()...)

	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	eng.Router().ServeHTTP(w, req)
	return w.Code
}

func TestScopeGateReadRequiresDirectoryRead(t *testing.T) {
	// Operator and auditor both hold directory:read → reads succeed.
	if code := serve(t, operator(), http.MethodGet, "/users", ""); code != http.StatusOK {
		t.Fatalf("operator GET /users = %d, want 200", code)
	}
	if code := serve(t, auditor(), http.MethodGet, "/users", ""); code != http.StatusOK {
		t.Fatalf("auditor GET /users = %d, want 200", code)
	}
	// An identity without directory:read is forbidden.
	none := scopedIdentity{scopes: map[string]bool{}}
	if code := serve(t, none, http.MethodGet, "/users", ""); code != http.StatusForbidden {
		t.Fatalf("scopeless GET /users = %d, want 403", code)
	}
}

func TestScopeGateMutationRequiresManageScope(t *testing.T) {
	body := `{"email":"jane@example.com","display_name":"Jane"}`

	// Operator holds users:manage → create succeeds (201).
	if code := serve(t, operator(), http.MethodPost, "/users", body); code != http.StatusCreated {
		t.Fatalf("operator POST /users = %d, want 201", code)
	}
	// Auditor (directory:read only) is forbidden from mutating — the exact
	// seeded auditor persona: reads everything, mutates nothing. The gate denies
	// before the handler runs, so these need no store wiring.
	denied := map[string]string{
		"POST /users":        "/users",        // users:manage
		"POST /tenants":      "/tenants",      // tenants:manage
		"POST /applications": "/applications", // applications:manage
	}
	for name, path := range denied {
		if code := serve(t, auditor(), http.MethodPost, path, "{}"); code != http.StatusForbidden {
			t.Fatalf("auditor %s = %d, want 403", name, code)
		}
	}
}
