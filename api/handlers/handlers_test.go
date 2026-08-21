package handlers

import (
	"context"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/janus/api/contracts"
	"github.com/zoobz-io/janus/api/wire"
	"github.com/zoobz-io/janus/database/models"
)

// Fakes satisfy the public contracts structurally so handlers can be driven
// through rocco's Process without a server or database. Error fields inject
// store failures; zero values behave as an empty store. Identity is rocco's
// NoIdentity (empty ID) — the fakes don't discriminate on it, and the
// WithAuthentication gate lives in the engine, not in Process.

type fakeUsers struct {
	user *models.User
	err  error
}

func (f *fakeUsers) GetUser(_ context.Context, _ string) (*models.User, error) {
	return f.user, f.err
}

func (f *fakeUsers) UpdateDisplayName(_ context.Context, _, _ string) (*models.User, error) {
	return f.user, f.err
}

type fakeMemberships struct{}

func (fakeMemberships) ListByUser(_ context.Context, _ string) ([]*models.Membership, error) {
	return nil, nil
}

func (fakeMemberships) ListByTenant(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Membership], error) {
	return &models.OffsetResult[models.Membership]{}, nil
}

func (fakeMemberships) GetByUserAndTenant(_ context.Context, _, _ string) (*models.Membership, error) {
	return nil, nil
}

func (fakeMemberships) Create(_ context.Context, _, _ string, _ models.UserRole) (*models.Membership, error) {
	return nil, nil
}

func (fakeMemberships) Delete(_ context.Context, _ string) error { return nil }

type fakeTenants struct{}

func (fakeTenants) GetTenant(_ context.Context, _ string) (*models.Tenant, error) {
	return nil, nil
}

func (fakeTenants) CreateTenant(_ context.Context, _, _ string) (*models.Tenant, error) {
	return nil, nil
}

type fakeSessions struct {
	sessions []*models.Session
	count    int
	err      error
}

func (f *fakeSessions) ListSessionsByUser(_ context.Context, _ string) ([]*models.Session, error) {
	return f.sessions, f.err
}

func (f *fakeSessions) RevokeSession(_ context.Context, _, _ string) error { return f.err }

func (f *fakeSessions) RevokeUserSessions(_ context.Context, _ string) (int, error) {
	return f.count, f.err
}

type fakeAccounts struct {
	accounts []*models.Account
	err      error
}

func (f *fakeAccounts) ListByUser(_ context.Context, _ string) ([]*models.Account, error) {
	return f.accounts, f.err
}

func (f *fakeAccounts) Unlink(_ context.Context, _, _ string) error { return f.err }

var (
	testUsers    = &fakeUsers{}
	testSessions = &fakeSessions{}
	testAccounts = &fakeAccounts{}
)

func TestMain(m *testing.M) {
	sum.New()
	k := sum.Start()
	wire.RegisterBoundaries(k)
	sum.Register[contracts.Users](k, testUsers)
	sum.Register[contracts.Memberships](k, fakeMemberships{})
	sum.Register[contracts.Tenants](k, fakeTenants{})
	sum.Register[contracts.Sessions](k, testSessions)
	sum.Register[contracts.Accounts](k, testAccounts)
	sum.Freeze(k)
	os.Exit(m.Run())
}

// process drives a handler through rocco's full request path and returns the
// HTTP status. pathValues populate Go 1.22 PathValue params (rocco's
// extractParams reads them from the request, no router required).
func process[In, Out any](t *testing.T, h *rocco.Handler[In, Out], method, target string, body io.Reader, pathValues map[string]string) int {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range pathValues {
		req.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	status, _ := h.Process(context.Background(), req, w)
	return status
}
