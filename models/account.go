package models

import "time"

// ProviderType identifies an external identity provider.
type ProviderType = string

// Supported identity providers.
const (
	ProviderZitadel ProviderType = "zitadel"
	ProviderAuth0   ProviderType = "auth0"
	ProviderGitHub  ProviderType = "github"
	ProviderGoogle  ProviderType = "google"
)

// Account maps an external IdP subject to an internal user.
// A single user may have multiple accounts across different
// providers — this enables federation so the same person can
// authenticate through any app on the mesh regardless of which IdP
// that app uses.
type Account struct {
	LinkedAt        time.Time    `json:"linked_at" db:"linked_at" default:"now()"`
	Provider        ProviderType `json:"provider" db:"provider" constraints:"notnull"`
	ExternalSubject string       `json:"external_subject" db:"external_subject" constraints:"notnull"`
	ID              string       `json:"id" db:"id" constraints:"primarykey"`
	UserID          string       `json:"user_id" db:"user_id" constraints:"notnull"`
}

// GetID returns the account's primary key.
func (l Account) GetID() string {
	return l.ID
}

// GetCreatedAt returns the account's creation timestamp.
func (l Account) GetCreatedAt() time.Time {
	return l.LinkedAt
}

// Clone returns a copy of the account.
func (l Account) Clone() Account {
	return l
}
