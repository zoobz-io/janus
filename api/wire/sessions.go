package wire

// SessionResponse is the public API response for a session.
type SessionResponse struct {
	ID        string `json:"id" description:"Session ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	IssuedBy  string `json:"issued_by" description:"Service that created the session" example:"argus"`
	UserAgent string `json:"user_agent,omitempty" description:"Device/browser info" example:"Mozilla/5.0"`
	IPAddress string `json:"ip_address,omitempty" description:"Last known IP" example:"192.168.1.1"`
	CreatedAt string `json:"created_at" description:"When the session was created" example:"2026-05-01T12:00:00Z"`
	ExpiresAt string `json:"expires_at" description:"When the session expires" example:"2026-05-02T12:00:00Z"`
}

// SessionListResponse is the public API response for a list of sessions.
type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions" description:"Active sessions"`
}

// Clone returns a deep copy of the response.
func (r SessionListResponse) Clone() SessionListResponse {
	c := r
	if r.Sessions != nil {
		c.Sessions = make([]SessionResponse, len(r.Sessions))
		copy(c.Sessions, r.Sessions)
	}
	return c
}

// RevokeAllSessionsResponse is the response for revoking all sessions.
type RevokeAllSessionsResponse struct {
	RevokedCount int `json:"revoked_count" description:"Number of sessions revoked" example:"3"`
}
