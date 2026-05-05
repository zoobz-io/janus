// Package auth provides session-based authentication for the janus public API.
package auth

import "github.com/zoobz-io/rocco"

// Compile-time assertion.
var _ rocco.Identity = (*JanusIdentity)(nil)

// JanusIdentity implements rocco.Identity backed by a janus user.
// Since users can belong to multiple tenants, TenantID returns empty —
// the calling handler determines tenant context from the request.
type JanusIdentity struct {
	userID string
	email  string
}

// ID returns the internal user ID.
func (j *JanusIdentity) ID() string { return j.userID }

// TenantID returns empty — users can belong to multiple tenants.
// Apps should determine tenant context from the request, not the identity.
func (j *JanusIdentity) TenantID() string { return "" }

// Email returns the user's email address.
func (j *JanusIdentity) Email() string { return j.email }

// Scopes returns nil — janus does not use scopes on the public API.
func (j *JanusIdentity) Scopes() []string { return nil }

// Roles returns nil — roles are per-tenant via memberships.
func (j *JanusIdentity) Roles() []string { return nil }

// HasScope returns false — janus does not use scopes on the public API.
func (j *JanusIdentity) HasScope(_ string) bool { return false }

// HasRole returns false — roles are per-tenant, not on the identity.
func (j *JanusIdentity) HasRole(_ string) bool { return false }

// Stats returns nil — no rate limiting stats on identity.
func (j *JanusIdentity) Stats() map[string]int { return nil }
