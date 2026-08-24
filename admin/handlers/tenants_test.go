package handlers

import "testing"

func TestTenantEndpointScopes(t *testing.T) {
	requireScope(t, listTenants, "directory:read")
	requireScope(t, searchTenants, "directory:read")
	requireScope(t, getTenant, "directory:read")
	requireScope(t, createTenant, "tenants:manage")
	requireScope(t, updateTenant, "tenants:manage")
}
