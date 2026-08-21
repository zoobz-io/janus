package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/zoobz-io/janus/database/models"
	"github.com/zoobz-io/janus/database/stores"
)

// A store miss maps to 404; any other store failure must surface as an
// internal error, not masquerade as not-found.
func TestGetMyProfile(t *testing.T) {
	t.Cleanup(func() { *testUsers = fakeUsers{} })

	t.Run("OK", func(t *testing.T) {
		*testUsers = fakeUsers{user: &models.User{ID: "u-1", Email: "a@b.c", DisplayName: "A"}}
		if got := process(t, getMyProfile, http.MethodGet, "/me", nil, nil); got != http.StatusOK {
			t.Fatalf("status = %d, want 200", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		*testUsers = fakeUsers{err: fmt.Errorf("user %w", stores.ErrNotFound)}
		if got := process(t, getMyProfile, http.MethodGet, "/me", nil, nil); got != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", got)
		}
	})

	t.Run("StoreErrorIsNot404", func(t *testing.T) {
		*testUsers = fakeUsers{err: errors.New("connection refused")}
		got := process(t, getMyProfile, http.MethodGet, "/me", nil, nil)
		if got == http.StatusNotFound {
			t.Fatal("store failure must not read as 404")
		}
		if got != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", got)
		}
	})
}

func TestUpdateMyProfile(t *testing.T) {
	t.Cleanup(func() { *testUsers = fakeUsers{} })
	body := func() *strings.Reader { return strings.NewReader(`{"display_name":"New Name"}`) }

	t.Run("OK", func(t *testing.T) {
		*testUsers = fakeUsers{user: &models.User{ID: "u-1", Email: "a@b.c", DisplayName: "New Name"}}
		if got := process(t, updateMyProfile, http.MethodPut, "/me", body(), nil); got != http.StatusOK {
			t.Fatalf("status = %d, want 200", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		*testUsers = fakeUsers{err: fmt.Errorf("user %w", stores.ErrNotFound)}
		if got := process(t, updateMyProfile, http.MethodPut, "/me", body(), nil); got != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", got)
		}
	})

	t.Run("StoreErrorIsNot404", func(t *testing.T) {
		*testUsers = fakeUsers{err: errors.New("connection refused")}
		if got := process(t, updateMyProfile, http.MethodPut, "/me", body(), nil); got != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", got)
		}
	})
}
