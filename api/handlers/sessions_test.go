package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/zoobz-io/janus/database/stores"
)

func TestRevokeMySession(t *testing.T) {
	t.Cleanup(func() { *testSessions = fakeSessions{} })
	id := map[string]string{"id": "sess-1"}

	t.Run("OK", func(t *testing.T) {
		*testSessions = fakeSessions{}
		if got := process(t, revokeMySession, http.MethodDelete, "/me/sessions/sess-1", nil, id); got != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		*testSessions = fakeSessions{err: fmt.Errorf("session %w", stores.ErrNotFound)}
		if got := process(t, revokeMySession, http.MethodDelete, "/me/sessions/sess-1", nil, id); got != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", got)
		}
	})

	t.Run("StoreErrorIsNot404", func(t *testing.T) {
		*testSessions = fakeSessions{err: errors.New("connection refused")}
		if got := process(t, revokeMySession, http.MethodDelete, "/me/sessions/sess-1", nil, id); got != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", got)
		}
	})
}
