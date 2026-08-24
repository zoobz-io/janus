package handlers

import "testing"

func TestTierEndpointScopes(t *testing.T) {
	requireScope(t, listTiers, "directory:read")
	requireScope(t, createTier, "applications:manage")
	requireScope(t, deleteTier, "applications:manage")
}
