package handlers

import "testing"

func TestUserEndpointScopes(t *testing.T) {
	requireScope(t, listUsers, "directory:read")
	requireScope(t, searchUsers, "directory:read")
	requireScope(t, getUser, "directory:read")
	requireScope(t, createUser, "users:manage")
	requireScope(t, updateUser, "users:manage")
}
