# internal/mesh

The gRPC service implementations that back the [mesh surface](../../cmd/mesh/) — four
[aegis](https://github.com/zoobz-io/aegis) services over the shared stores. Peer services
call these to resolve and register identities, mint and validate sessions, read the
directory, and manage entitlements. Auth is the mesh CA and the client cert; the calling
application is **its cert CN**.

## The calling application

Everything entitlement-aware here starts by identifying the caller. `callerAppSlug`
([`entitlement.go`](entitlement.go)) reads `sc.Metadata.NodeID` from the aegis security
context — the client cert CN — which maps to `Application.Slug`. When that slug is
unresolvable the server does **not** fail hard: it warns `EntitlementCheckSkipped` and
returns without tenants. `entitlementChecker` adapts the shared
[`authz.Entitlements`](../authz/) resolver to the protobuf surface.

## DirectoryServer — [`directory.go`](directory.go)

`GetUser`, `GetUserByEmail`, `GetTenant`, `CreateTenant`, `UpdateTenant`. Holds the full
store aggregate (not just individual stores) because `CreateTenant` with an owner is a
transactional multi-store flow: `CreateTenantWithOwner` lands the tenant and the owner
membership in one transaction, and `TenantCreated` emits **only post-commit** — a
rolled-back create emits nothing. Without an owner it's a plain `CreateTenant`.

## IdentityServer — [`identity.go`](identity.go)

- `ResolveIdentity` maps `provider`+`subject` → user via
  `accounts.GetByProviderSubject`, touches last-seen (a failure there is a warning, not a
  refusal), then **scopes the result to the calling app's entitlement** — no entitled
  tenants means the user isn't entitled to that caller, and it errors.
- `Register` creates user + identity + optional tenant in **one transaction**, emitting
  `UserCreated` / `IdentityLinked` / `TenantCreated` only post-commit.
- `ListProviders` lists the external providers linked to a user.

## SessionServer — [`session.go`](session.go)

- `CreateSession` issues a token; the session's `issuedBy` is the **aegis caller's
  NodeID** (`"unknown"` if there's no caller context).
- `ValidateSession` validates the token **then** scopes it to the calling app's
  entitlement. A good token with **no entitled tenants for that caller** returns
  `Valid=false` — the session is valid globally but not *for this app*. This is the
  one-round-trip authenticate-and-authorize.
- `RevokeSession`, `RevokeUserSessions`, and `ListUserSessions` (which returns 8-char
  token prefixes, never full tokens).
- `SubscribeSessionEvents` returns **"not yet implemented"** — the streaming RPC is
  stubbed.

## EntitlementServer — [`entitlement.go`](entitlement.go)

Manages licenses and grants **for the calling app** (resolved by cert CN via
`resolveApp`): `AuthorizeApplication` / `RevokeApplication`, `GrantUserAccess` /
`RevokeUserAccess` / `UpdateUserAccess`, and the `List*` reads. Every **mutation** gates
on `requireAdmin` → `authz.RequireRole(... admin, owner)` in the target tenant; the reads
(`ListTenantApplications`, `ListUserAccess`) do not.
