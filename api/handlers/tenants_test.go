package handlers

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/zoobz-io/janus/database/models"
)

func TestCreateMyTenant(t *testing.T) {
	t.Cleanup(func() { *testProvisioning = fakeProvisioning{} })
	body := func() *strings.Reader { return strings.NewReader(`{"name":"Acme Corp","slug":"acme-corp"}`) }

	t.Run("OK", func(t *testing.T) {
		*testProvisioning = fakeProvisioning{tenant: &models.Tenant{ID: "t-1", Name: "Acme Corp", Slug: "acme-corp"}}
		if got := process(t, createMyTenant, http.MethodPost, "/me/tenants", body(), nil); got != http.StatusCreated {
			t.Fatalf("status = %d, want 201", got)
		}
	})

	t.Run("ProvisioningError", func(t *testing.T) {
		*testProvisioning = fakeProvisioning{err: errors.New("connection refused")}
		if got := process(t, createMyTenant, http.MethodPost, "/me/tenants", body(), nil); got != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", got)
		}
	})
}
