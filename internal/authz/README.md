# internal/authz

Portable authorization for tenant-scoped operations. It answers "may this user do this
here?" from both HTTP handlers and gRPC servers, because it depends on **small
structural interfaces over the stores** — never on an API contract package. The concrete
stores satisfy those interfaces; so do the public and admin membership contracts. That's
the whole point: one authorization core, no coupling to any single surface.

Two concerns live here.

## Role checks

[`authz.go`](authz.go), errors in [`errors.go`](errors.go).

- `RequireRole(ctx, memberships, userID, tenantID, roles...)` returns the membership on
  success, or `ErrNotMember` / `ErrInsufficientRole`. A store failure is returned
  **as-is** — an infrastructure error must never masquerade as an authorization denial.
- `RequireOwnerExists(ctx, memberships, tenantID, excludeUserID)` guards the last owner:
  `ErrLastOwner` when demoting or removing the excluded user would leave the tenant
  ownerless. It counts *other* owners via `CountOtherOwners`.

The mesh's `EntitlementServer` gates every mutation through `RequireRole(... admin,
owner)` — see [`internal/mesh`](../mesh/).

## Entitlements

[`entitlements.go`](entitlements.go) — the `Entitlements` resolver, and the one
resolution path shared by the mesh and the public API.

`ForApplication(ctx, userID, appSlug)` resolves the app by slug, then walks the user's
memberships and includes a tenant **only if it has both**:

- a **license** (`GetByTenantAndApp` — the tenant is authorized for the app), **and**
- a **grant** (`GetByUserAndApp` — the user has access within that tenant).

Miss either and the tenant is skipped. The result is `[]models.AuthorizedTenant`, each
carrying the user's tenant `Role`, the grant's app-defined `AppRoles`, and the
`AppScopes` — the user's **effective scopes** for the app.

Effective scopes (`effectiveScopes`) are the union of the grant's **explicit** scopes and
the scopes bundled into the grant's **tier's features** (`features.ListByTier` →
`scopes.GetByID`), sorted. Resolved **live** on every call, so editing a tier takes
effect immediately — no cached entitlement to invalidate.

An unentitled user yields an **empty list, not an error**. An unknown slug returns an
error wrapping `ErrApplicationNotFound`.

This resolver backs `/me/authorization/{app_slug}` on the [public API](../../api/) and
`ValidateSession` on the [mesh](../mesh/) — the same walk, one authenticated round-trip
away from also authorizing a request.
