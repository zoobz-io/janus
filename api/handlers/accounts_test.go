package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/zoobz-io/janus/database/stores"
)

func TestUnlinkMyAccount(t *testing.T) {
	t.Cleanup(func() { *testAccounts = fakeAccounts{} })
	id := map[string]string{"id": "acct-1"}

	t.Run("OK", func(t *testing.T) {
		*testAccounts = fakeAccounts{}
		if got := process(t, unlinkMyAccount, http.MethodDelete, "/me/accounts/acct-1", nil, id); got != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		*testAccounts = fakeAccounts{err: fmt.Errorf("account %w", stores.ErrNotFound)}
		if got := process(t, unlinkMyAccount, http.MethodDelete, "/me/accounts/acct-1", nil, id); got != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", got)
		}
	})

	t.Run("StoreErrorIsNot404", func(t *testing.T) {
		*testAccounts = fakeAccounts{err: errors.New("connection refused")}
		if got := process(t, unlinkMyAccount, http.MethodDelete, "/me/accounts/acct-1", nil, id); got != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", got)
		}
	})
}
