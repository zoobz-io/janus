package events

import "github.com/zoobz-io/capitan"

// Domain event signals for identity lifecycle.
var (
	UserCreated    = capitan.NewSignal("janus.user.created", "New user provisioned")
	UserUpdated    = capitan.NewSignal("janus.user.updated", "User profile updated")
	SessionCreated = capitan.NewSignal("janus.session.created", "Session token issued")
	SessionRevoked = capitan.NewSignal("janus.session.revoked", "Session token revoked")
	SessionExpired = capitan.NewSignal("janus.session.expired", "Session token expired")
	IdentityLinked   = capitan.NewSignal("janus.identity.linked", "External identity linked to user")
	IdentityUnlinked = capitan.NewSignal("janus.identity.unlinked", "External identity unlinked from user")
	TenantCreated = capitan.NewSignal("janus.tenant.created", "New tenant provisioned")
	TenantUpdated = capitan.NewSignal("janus.tenant.updated", "Tenant updated")
)

// Domain event field keys.
var (
	UserIDKey     = capitan.NewStringKey("user_id")
	TenantIDKey   = capitan.NewStringKey("tenant_id")
	SessionIDKey  = capitan.NewStringKey("session_id")
	IdentityIDKey = capitan.NewStringKey("identity_id")
	ProviderKey   = capitan.NewStringKey("provider")
	IssuedByKey   = capitan.NewStringKey("issued_by")
	EmailKey      = capitan.NewStringKey("email")
)
