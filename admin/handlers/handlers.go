package handlers

import "github.com/zoobz-io/rocco"

// All returns the admin API endpoints. Access requires a janus-admin operator
// entitlement (enforced by the admin authenticator), and each endpoint declares
// the operator scope it needs via WithScopes: directory:read for reads,
// users:manage / tenants:manage / applications:manage for the respective
// mutations. An auditor (directory:read only) can read everything and mutate
// nothing.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Applications
		listApplications, searchApplications, getApplication, createApplication, updateApplication,
		// Tenants
		listTenants, searchTenants, getTenant, createTenant, updateTenant,
		// Tenant members
		listMembers, addMember, updateMemberRole, removeMember,
		// Users
		listUsers, searchUsers, getUser, createUser, updateUser,
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
