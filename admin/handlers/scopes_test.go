package handlers

import "testing"

func TestScopeEndpointScopes(t *testing.T) {
	requireScope(t, listScopes, "directory:read")
	requireScope(t, createScope, "applications:manage")
	requireScope(t, deleteScope, "applications:manage")
}
