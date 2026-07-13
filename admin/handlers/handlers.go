package handlers

import "github.com/zoobz-io/rocco"

// All returns the admin API endpoints. Authentication is identical to the
// public API — any authenticated session is accepted; a finer operator
// authorization model is layered on separately.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Applications
		listApplications, getApplication, createApplication, updateApplication,
		// Tenants
		listTenants, getTenant, createTenant, updateTenant,
		// Tenant members
		listMembers, addMember, updateMemberRole, removeMember,
		// Users
		listUsers, getUser, createUser, updateUser,
		// User sessions
		listUserSessions, revokeUserSession, revokeAllUserSessions,
		// User accounts
		listUserAccounts, unlinkUserAccount,
		// Licenses
		listLicenses, authorizeLicense, revokeLicense,
		// Grants
		listGrants, createGrant, updateGrant, revokeGrant,
		// Scopes
		listScopes, createScope, deleteScope,
		// Tiers
		listTiers, createTier, deleteTier,
		// Features
		listFeatures, addFeature, removeFeature,
		// Providers
		listProviders,
	}
}
