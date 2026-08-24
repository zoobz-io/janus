package handlers

import "testing"

func TestLicenseEndpointScopes(t *testing.T) {
	requireScope(t, listLicenses, "directory:read")
	requireScope(t, authorizeLicense, "applications:manage")
	requireScope(t, revokeLicense, "applications:manage")
}
