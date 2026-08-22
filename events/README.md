# Events

Typed [capitan](https://github.com/zoobz-io/capitan) events with `janus.*` signals,
bridged to OpenTelemetry via [aperture](https://github.com/zoobz-io/aperture) — the metric
and log mapping lives in the `aperture_schema` config row polled by
[`internal/observe`](../internal/observe/). Domain mutations emit; the bridge turns the
emissions into counters and logs.

Three files, three flavors.

## Domain events — [`domain.go`](domain.go)

Flat package-level vars, each declared with
`sum.NewInfoEvent[Payload](sum.NewSignal("janus.…", "description"))`. Eighteen of them,
grouped by the payload struct they carry.

| Payload | Fields | Events |
|---------|--------|--------|
| `UserEvent` | `UserID`, `Email` | `UserCreated`, `UserUpdated` |
| `SessionEvent` | `SessionID`, `UserID`, `IssuedBy` | `SessionCreated`, `SessionRevoked`, `SessionExpired` |
| `IdentityEvent` | `UserID`, `Provider` | `IdentityLinked`, `IdentityUnlinked` |
| `TenantEvent` | `TenantID` | `TenantCreated`, `TenantUpdated` |
| `MembershipEvent` | `UserID`, `TenantID`, `Role` | `MemberAdded`, `MemberUpdated`, `MemberRemoved` |
| `AppEntitlementEvent` | `TenantID`, `ApplicationID`, `UserID` (empty for tenant-level) | `TenantAppAuthorized`, `TenantAppRevoked`, `UserAppGranted`, `UserAppRevoked` |
| `ApplicationEvent` | `ApplicationID`, `Name` | `ApplicationCreated`, `ApplicationUpdated` |

Signal names follow the payload's domain: `janus.user.created`, `janus.session.revoked`,
`janus.tenant.app.authorized`, and so on.

Most of these exist to be counted (see the metric list in
[`migrations/002`](../database/migrations/002_aperture_config.sql)), but two drive real
machinery: `ApplicationCreated` and `ApplicationUpdated` feed the id↔name label cache in
[`internal/labels`](../internal/labels/) — its listener writes the new `Name` into Redis
on every emission.

## Operational signals — [`operational.go`](operational.go)

Raw `capitan.NewSignal` warnings, not typed events — they carry key/value context rather
than a domain payload. Three:

- `EntitlementCheckSkipped` — `janus.ops.entitlement.skipped`
- `LastSeenUpdateFailed` — `janus.ops.last_seen.failed`
- `LabelSyncFailed` — `janus.ops.label_sync.failed` (emitted when the labels listener
  above fails to write a mapping)

Field keys `OpUserIDKey`, `OpAppIDKey`, `OpErrorKey` accompany them.

## Startup signals — [`startup.go`](startup.go)

`janus.startup.*` lifecycle markers, also raw capitan signals: `DatabaseConnected`,
`RedisConnected`, `ServicesReady`, `OTELReady`, `ApertureReady`, `ServerListening`,
`MeshReady`. Field keys `StartupPortKey`, `StartupWorkersKey`, `StartupErrorKey`.
