package handlers

import "github.com/zoobz-io/rocco"

// All returns the authenticated API endpoints.
// Auth handlers (login/callback/logout) are added separately via AllWithAuth
// since they require runtime configuration.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Profile
		getMyProfile,
		updateMyProfile,
		// Tenants
		createMyTenant,
		// Sessions
		listMySessions,
		revokeMySession,
		revokeAllMySessions,
		// Accounts
		listMyAccounts,
		unlinkMyAccount,
		// Applications (read-only)
		listApplications,
		listMyApplications,
	}
}

// AllWithAuth returns All() plus the auth handlers (login, callback, logout).
func AllWithAuth(auth *AuthHandlers) []rocco.Endpoint {
	endpoints := All()
	if auth != nil {
		endpoints = append(endpoints, auth.Login, auth.Callback, auth.Logout)
	}
	return endpoints
}
