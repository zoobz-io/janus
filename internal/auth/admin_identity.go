package auth

import (
	"sort"

	"github.com/zoobz-io/rocco"

	"github.com/zoobz-io/janus/database/models"
)

// Compile-time assertion.
var _ rocco.Identity = (*AdminIdentity)(nil)

// AdminIdentity implements rocco.Identity with real operator scopes and roles,
// resolved from the caller's janus-admin entitlements. Unlike JanusIdentity
// (public API, unscoped), it carries the union of application scopes/roles across
// every tenant that entitles the user to janus-admin — so rocco's WithScopes can
// gate each admin endpoint.
type AdminIdentity struct {
	scopeSet map[string]struct{}
	roleSet  map[string]struct{}
	userID   string
	email    string
	scopes   []string
	roles    []string
}

// NewAdminIdentity builds an admin identity from the tenants that entitle the
// user to janus-admin, unioning their app scopes and roles.
func NewAdminIdentity(userID, email string, tenants []models.AuthorizedTenant) *AdminIdentity {
	scopeSet := map[string]struct{}{}
	roleSet := map[string]struct{}{}
	for _, t := range tenants {
		for _, s := range t.AppScopes {
			scopeSet[s] = struct{}{}
		}
		for _, r := range t.AppRoles {
			roleSet[r] = struct{}{}
		}
	}
	return &AdminIdentity{
		userID:   userID,
		email:    email,
		scopes:   sortedKeys(scopeSet),
		roles:    sortedKeys(roleSet),
		scopeSet: scopeSet,
		roleSet:  roleSet,
	}
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ID returns the internal user ID.
func (a *AdminIdentity) ID() string { return a.userID }

// TenantID returns empty — admin operates across tenants, not within one.
func (a *AdminIdentity) TenantID() string { return "" }

// Email returns the user's email address.
func (a *AdminIdentity) Email() string { return a.email }

// Scopes returns the union of the user's janus-admin scopes.
func (a *AdminIdentity) Scopes() []string { return a.scopes }

// Roles returns the union of the user's janus-admin application roles.
func (a *AdminIdentity) Roles() []string { return a.roles }

// HasScope reports whether the identity holds the given scope.
func (a *AdminIdentity) HasScope(scope string) bool {
	_, ok := a.scopeSet[scope]
	return ok
}

// HasRole reports whether the identity holds the given role.
func (a *AdminIdentity) HasRole(role string) bool {
	_, ok := a.roleSet[role]
	return ok
}

// Stats returns nil — no rate-limiting stats on identity.
func (a *AdminIdentity) Stats() map[string]int { return nil }
