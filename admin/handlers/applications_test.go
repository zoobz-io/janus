package handlers

import "testing"

func TestApplicationEndpointScopes(t *testing.T) {
	requireScope(t, listApplications, "directory:read")
	requireScope(t, getApplication, "directory:read")
	requireScope(t, searchApplications, "directory:read")
	requireScope(t, createApplication, "applications:manage")
	requireScope(t, updateApplication, "applications:manage")
}
