package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/zoobz-io/rocco"
	roccotest "github.com/zoobz-io/rocco/testing"
)

// TestApplicationEndpointsEnforceScope verifies each endpoint denies a caller that
// holds every admin scope except the one it requires — pinning the exact scope.
// The scope gate denies before the handler runs, so no store wiring is needed.
func TestApplicationEndpointsEnforceScope(t *testing.T) {
	// Reads require directory:read.
	readID := roccotest.NewMockIdentity("op").WithScopes("users:manage", "tenants:manage", "applications:manage")
	reads := roccotest.TestEngineWithAuth(func(context.Context, *http.Request) (rocco.Identity, error) { return readID, nil }).WithHandlers(All()...)
	roccotest.AssertStatus(t, roccotest.ServeRequest(reads, "GET", "/applications", nil), http.StatusForbidden)
	roccotest.AssertStatus(t, roccotest.ServeRequest(reads, "GET", "/applications/a1", nil), http.StatusForbidden)
	roccotest.AssertStatus(t, roccotest.ServeRequest(reads, "POST", "/applications/search", nil), http.StatusForbidden)

	// Mutations require their manage scope.
	writeID := roccotest.NewMockIdentity("op").WithScopes("directory:read", "users:manage", "tenants:manage")
	writes := roccotest.TestEngineWithAuth(func(context.Context, *http.Request) (rocco.Identity, error) { return writeID, nil }).WithHandlers(All()...)
	roccotest.AssertStatus(t, roccotest.ServeRequest(writes, "POST", "/applications", nil), http.StatusForbidden)
	roccotest.AssertStatus(t, roccotest.ServeRequest(writes, "PATCH", "/applications/a1", nil), http.StatusForbidden)
}
