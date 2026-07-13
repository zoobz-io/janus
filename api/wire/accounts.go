package wire

// AccountResponse is the public API response for a account provider.
type AccountResponse struct {
	ID              string `json:"id" description:"Account ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	Provider        string `json:"provider" description:"Identity provider" example:"zitadel"`
	ExternalSubject string `json:"external_subject" description:"External subject ID" example:"abc123"`
	LinkedAt        string `json:"linked_at" description:"When the identity was linked" example:"2026-05-01T12:00:00Z"`
}

// AccountListResponse is the public API response for a list of accounts.
type AccountListResponse struct {
	Accounts []AccountResponse `json:"accounts" description:"Account providers"`
}

// Clone returns a deep copy of the response.
func (r AccountListResponse) Clone() AccountListResponse {
	c := r
	if r.Accounts != nil {
		c.Accounts = make([]AccountResponse, len(r.Accounts))
		copy(c.Accounts, r.Accounts)
	}
	return c
}
