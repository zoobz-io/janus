package models

// AuthorizedTenant is the resolved view of a user's access to one application
// through one tenant: the tenant, the user's membership role in it, and the
// application-specific roles and effective scopes from the user's grant. It is
// computed from License + Grant (+ tier features) at resolution time, never
// persisted.
type AuthorizedTenant struct {
	TenantID   string   `json:"tenant_id"`
	TenantName string   `json:"tenant_name"`
	Role       UserRole `json:"role"`
	AppRoles   []string `json:"app_roles"`
	AppScopes  []string `json:"app_scopes"`
}
