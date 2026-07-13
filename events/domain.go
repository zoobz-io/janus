package events

import "github.com/zoobz-io/sum"

// UserEvent carries data for user lifecycle events.
type UserEvent struct {
	UserID string
	Email  string
}

// SessionEvent carries data for session lifecycle events.
type SessionEvent struct {
	SessionID string
	UserID    string
	IssuedBy  string
}

// IdentityEvent carries data for identity link/unlink events.
type IdentityEvent struct {
	UserID   string
	Provider string
}

// TenantEvent carries data for tenant lifecycle events.
type TenantEvent struct {
	TenantID string
}

// MembershipEvent carries data for membership changes.
type MembershipEvent struct {
	UserID   string
	TenantID string
	Role     string
}

// AppEntitlementEvent carries data for application entitlement changes.
type AppEntitlementEvent struct {
	TenantID      string
	ApplicationID string
	UserID        string // empty for tenant-level events
}

// Domain events for identity lifecycle.
var (
	UserCreated         = sum.NewInfoEvent[UserEvent](sum.NewSignal("janus.user.created", "New user provisioned"))
	UserUpdated         = sum.NewInfoEvent[UserEvent](sum.NewSignal("janus.user.updated", "User profile updated"))
	SessionCreated      = sum.NewInfoEvent[SessionEvent](sum.NewSignal("janus.session.created", "Session token issued"))
	SessionRevoked      = sum.NewInfoEvent[SessionEvent](sum.NewSignal("janus.session.revoked", "Session token revoked"))
	SessionExpired      = sum.NewInfoEvent[SessionEvent](sum.NewSignal("janus.session.expired", "Session token expired"))
	IdentityLinked      = sum.NewInfoEvent[IdentityEvent](sum.NewSignal("janus.identity.linked", "External identity linked to user"))
	IdentityUnlinked    = sum.NewInfoEvent[IdentityEvent](sum.NewSignal("janus.identity.unlinked", "External identity unlinked from user"))
	TenantCreated       = sum.NewInfoEvent[TenantEvent](sum.NewSignal("janus.tenant.created", "New tenant provisioned"))
	TenantUpdated       = sum.NewInfoEvent[TenantEvent](sum.NewSignal("janus.tenant.updated", "Tenant updated"))
	MemberAdded         = sum.NewInfoEvent[MembershipEvent](sum.NewSignal("janus.member.added", "Member added to tenant"))
	MemberUpdated       = sum.NewInfoEvent[MembershipEvent](sum.NewSignal("janus.member.updated", "Member role updated"))
	MemberRemoved       = sum.NewInfoEvent[MembershipEvent](sum.NewSignal("janus.member.removed", "Member removed from tenant"))
	TenantAppAuthorized = sum.NewInfoEvent[AppEntitlementEvent](sum.NewSignal("janus.tenant.app.authorized", "Tenant authorized for application"))
	TenantAppRevoked    = sum.NewInfoEvent[AppEntitlementEvent](sum.NewSignal("janus.tenant.app.revoked", "Tenant authorization for application revoked"))
	UserAppGranted      = sum.NewInfoEvent[AppEntitlementEvent](sum.NewSignal("janus.user.app.granted", "User granted access to application"))
	UserAppRevoked      = sum.NewInfoEvent[AppEntitlementEvent](sum.NewSignal("janus.user.app.revoked", "User access to application revoked"))
)
