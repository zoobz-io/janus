package handlers

import "testing"

func TestMemberEndpointScopes(t *testing.T) {
	requireScope(t, listMembers, "directory:read")
	requireScope(t, addMember, "tenants:manage")
	requireScope(t, updateMemberRole, "tenants:manage")
	requireScope(t, removeMember, "tenants:manage")
}
