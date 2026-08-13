package wire

// AuthorizedTenantResponse is one tenant through which the caller holds
// access to an application.
type AuthorizedTenantResponse struct {
	TenantID   string   `json:"tenant_id" description:"Tenant ID" example:"660e8400-e29b-41d4-a716-446655440000"`
	TenantName string   `json:"tenant_name" description:"Tenant name" example:"Acme Corp"`
	Role       string   `json:"role" description:"The caller's membership role in this tenant" example:"admin"`
	Roles      []string `json:"roles" description:"Application roles on the caller's grant"`
	Scopes     []string `json:"scopes" description:"Effective scopes: grant scopes unioned with tier features"`
}

// AuthorizationResponse is the caller's resolved authorization for one
// application — the payload a consumer app authenticates its users with.
type AuthorizationResponse struct {
	UserID      string                     `json:"user_id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	Application ApplicationResponse        `json:"application" description:"The application authorization was resolved for"`
	Tenants     []AuthorizedTenantResponse `json:"tenants" description:"Tenants entitling the caller to the application; empty when the caller has no access"`
}

// Clone returns a deep copy of the response.
func (r AuthorizationResponse) Clone() AuthorizationResponse {
	c := r
	if r.Tenants != nil {
		c.Tenants = make([]AuthorizedTenantResponse, len(r.Tenants))
		for i, t := range r.Tenants {
			ct := t
			if t.Roles != nil {
				ct.Roles = make([]string, len(t.Roles))
				copy(ct.Roles, t.Roles)
			}
			if t.Scopes != nil {
				ct.Scopes = make([]string, len(t.Scopes))
				copy(ct.Scopes, t.Scopes)
			}
			c.Tenants[i] = ct
		}
	}
	return c
}
