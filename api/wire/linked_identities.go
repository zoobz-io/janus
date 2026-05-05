package wire

// LinkedIdentityResponse is the public API response for a linked identity provider.
type LinkedIdentityResponse struct {
	ID              string `json:"id" description:"Linked identity ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	Provider        string `json:"provider" description:"Identity provider" example:"zitadel"`
	ExternalSubject string `json:"external_subject" description:"External subject ID" example:"abc123"`
	LinkedAt        string `json:"linked_at" description:"When the identity was linked" example:"2026-05-01T12:00:00Z"`
}

// LinkedIdentityListResponse is the public API response for a list of linked identities.
type LinkedIdentityListResponse struct {
	Identities []LinkedIdentityResponse `json:"identities" description:"Linked identity providers"`
}

// Clone returns a deep copy of the response.
func (r LinkedIdentityListResponse) Clone() LinkedIdentityListResponse {
	c := r
	if r.Identities != nil {
		c.Identities = make([]LinkedIdentityResponse, len(r.Identities))
		copy(c.Identities, r.Identities)
	}
	return c
}
