package wire

// ProviderListResponse is the admin API response for the supported identity providers.
type ProviderListResponse struct {
	Providers []string `json:"providers" description:"Supported identity providers" example:"zitadel"`
}

// Clone returns a deep copy of the response.
func (r ProviderListResponse) Clone() ProviderListResponse {
	c := r
	if r.Providers != nil {
		c.Providers = make([]string, len(r.Providers))
		copy(c.Providers, r.Providers)
	}
	return c
}
