package events

import "github.com/zoobz-io/capitan"

// Operational signals for non-fatal warnings and diagnostics.
// These are raw capitan signals (not sum.Event) since they carry
// simple key-value context rather than typed domain data.
var (
	EntitlementCheckSkipped = capitan.NewSignal("janus.ops.entitlement.skipped", "Entitlement check skipped — no security context")
	LastSeenUpdateFailed    = capitan.NewSignal("janus.ops.last_seen.failed", "Failed to update user last_seen timestamp")
)

// Operational field keys.
var (
	OpUserIDKey = capitan.NewStringKey("user_id")
	OpErrorKey  = capitan.NewErrorKey("error")
)
