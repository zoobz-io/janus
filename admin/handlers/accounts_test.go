package handlers

import "testing"

func TestAccountEndpointScopes(t *testing.T) {
	requireScope(t, listUserAccounts, "directory:read")
	requireScope(t, unlinkUserAccount, "users:manage")
}
