package handlers

import "testing"

func TestProviderEndpointScopes(t *testing.T) {
	requireScope(t, listProviders, "directory:read")
}
