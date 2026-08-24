package handlers

import "testing"

func TestGrantEndpointScopes(t *testing.T) {
	requireScope(t, listGrants, "directory:read")
	requireScope(t, createGrant, "applications:manage")
	requireScope(t, updateGrant, "applications:manage")
	requireScope(t, revokeGrant, "applications:manage")
}
