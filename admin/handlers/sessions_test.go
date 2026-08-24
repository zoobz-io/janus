package handlers

import "testing"

func TestSessionEndpointScopes(t *testing.T) {
	requireScope(t, listUserSessions, "directory:read")
	requireScope(t, revokeUserSession, "users:manage")
	requireScope(t, revokeAllUserSessions, "users:manage")
}
