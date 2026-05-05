package handlers

import "github.com/zoobz-io/rocco"

// All returns the complete set of public API endpoints.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Profile
		getMyProfile,
		updateMyProfile,
		// Sessions
		listMySessions,
		revokeMySession,
		revokeAllMySessions,
		// Linked Identities
		listMyIdentities,
		unlinkMyIdentity,
	}
}
