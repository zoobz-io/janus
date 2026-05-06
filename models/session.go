package models

import "time"

// Session represents an active authentication session.
// Sessions are mesh-wide — a token issued by one app is valid
// across all apps on the mesh. The IssuedBy field records which
// service created the session.
//
// The raw token is never stored — only a SHA-256 hash. Since the
// hash is already irreversible, no boundary encryption is needed.
type Session struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at" constraints:"notnull"`
	TokenHash string    `json:"-" db:"token_hash" constraints:"notnull,unique"`
	ID        string    `json:"id" db:"id" constraints:"primarykey"`
	UserID    string    `json:"user_id" db:"user_id" constraints:"notnull"`
	IssuedBy  string    `json:"issued_by" db:"issued_by" constraints:"notnull"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
}

// GetID returns the session's primary key.
func (s Session) GetID() string {
	return s.ID
}

// GetCreatedAt returns the session's creation timestamp.
func (s Session) GetCreatedAt() time.Time {
	return s.CreatedAt
}

// Expired returns true if the session has passed its expiration time.
func (s Session) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Clone returns a copy of the session.
func (s Session) Clone() Session {
	return s
}
