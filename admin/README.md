# Admin API

The operator HTTP surface, served by [`cmd/admin`](../cmd/admin/) on `APP_ADMIN_PORT`
(default 8081). Where the [public API](../api/) is scoped to the caller's own identity,
the admin API is scoped to *everything*: register applications and their scope/tier
catalogs, license tenants, grant users, manage members, revoke sessions. It is the
write side of the whole domain, and the [web console](../web/) is its UI.

It follows the same [four-package pattern as the public API](../api/#the-four-package-pattern)
— `contracts → handlers → transformers → wire`. The only structural difference is the
capability boundary: admin's contract interfaces expose the broad methods (create,
update, list-all-statuses) that the public contracts withhold, over the *same* store
instances registered in [`cmd/admin/main.go`](../cmd/admin/main.go).

## Authentication and authorization

Every endpoint calls `.WithAuthentication()`, and the admin binary installs the *same*
session authenticator as the public API — a signed session cookie or a bearer token,
validated against the shared sessions store. A browser session minted by the public
login flow is accepted here because both surfaces share the session store and cookie key;
a same-origin proxy forwards the cookie from the console.

**Known gap — read before you deploy this open.** Authentication is the only gate. There
is no per-endpoint *authorization*: any authenticated session may call any admin endpoint.
The `operator` role and its scopes (`directory:read`, `users:manage`, `tenants:manage`,
`applications:manage`) exist in seed data and in the web console's contract, but nothing
in this package enforces them yet. The one authorization rule that *is* enforced is
last-owner protection on tenant membership (`authz.RequireOwnerExists` in
[`handlers/members.go`](handlers/members.go)). Until an operator authorization model
lands, gate this surface at the network edge. Tracked as follow-up work.

## Endpoints

33 endpoints, registered in [`handlers/handlers.go`](handlers/handlers.go) and grouped by
the two OpenAPI tag groups: **Catalog** (what an application offers and who may use it)
and **Directory** (organizations and people).

### Catalog

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/applications` | List all applications, including inactive |
| `POST` | `/applications/search` | Search apps (text, status facet, date ranges, sort, page) |
| `GET` | `/applications/{app_id}` | Get one application |
| `POST` | `/applications` | Register an application; slug must match the service cert CN |
| `PATCH` | `/applications/{app_id}` | Update name/status (`inactive` soft-disables) |
| `GET` | `/applications/{app_id}/scopes` | List an app's scope catalog |
| `POST` | `/applications/{app_id}/scopes` | Define a scope |
| `DELETE` | `/applications/{app_id}/scopes/{scope_id}` | Delete a scope (features cascade) |
| `GET` | `/applications/{app_id}/tiers` | List tiers, ordered by rank |
| `POST` | `/applications/{app_id}/tiers` | Define a tier |
| `DELETE` | `/applications/{app_id}/tiers/{tier_id}` | Delete a tier (features cascade) |
| `GET` | `/applications/{app_id}/tiers/{tier_id}/features` | List scopes bundled into a tier |
| `POST` | `/applications/{app_id}/tiers/{tier_id}/features` | Bundle a scope into a tier |
| `DELETE` | `/applications/{app_id}/tiers/{tier_id}/features/{scope_id}` | Unbundle a scope |
| `GET` | `/applications/{app_id}/licenses` | List tenants licensed for an app |
| `POST` | `/applications/{app_id}/licenses` | License a tenant for an app |
| `DELETE` | `/applications/{app_id}/licenses/{tenant_id}` | Revoke a tenant's license |
| `GET` | `/applications/{app_id}/grants?tenant_id=` | List per-user grants for an app in a tenant |
| `POST` | `/applications/{app_id}/grants` | Grant a user access (roles/scopes + optional tier) |
| `PATCH` | `/applications/{app_id}/grants/{tenant_id}/{user_id}` | Replace roles/scopes, set/clear tier |
| `DELETE` | `/applications/{app_id}/grants/{tenant_id}/{user_id}` | Revoke a user's grant |

### Directory

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/tenants?limit=&offset=` | List tenants (offset-paginated) |
| `POST` | `/tenants/search` | Search tenants |
| `GET` | `/tenants/{tenant_id}` | Get a tenant |
| `POST` | `/tenants` | Create a tenant |
| `PATCH` | `/tenants/{tenant_id}` | Update name/status (`suspended` disables) |
| `GET` | `/tenants/{tenant_id}/members?limit=&offset=` | List members with roles |
| `POST` | `/tenants/{tenant_id}/members` | Add a member with a role |
| `PATCH` | `/tenants/{tenant_id}/members/{user_id}` | Change a role (last-owner demotion rejected) |
| `DELETE` | `/tenants/{tenant_id}/members/{user_id}` | Remove a member (last-owner removal rejected) |
| `GET` | `/users?email=&limit=&offset=` | List users (`email` narrows to one) |
| `POST` | `/users/search` | Search users |
| `GET` | `/users/{user_id}` | Get a user |
| `POST` | `/users` | Create a user directly (normally OIDC-provisioned) |
| `PATCH` | `/users/{user_id}` | Update display name/status |
| `GET` | `/users/{user_id}/sessions` | List a user's active sessions |
| `DELETE` | `/users/{user_id}/sessions/{session_id}` | Revoke one session |
| `DELETE` | `/users/{user_id}/sessions` | Revoke all sessions, returns count |
| `GET` | `/users/{user_id}/accounts` | List linked identity accounts |
| `DELETE` | `/users/{user_id}/accounts/{account_id}` | Unlink an account |
| `GET` | `/providers` | List supported identity providers (static) |

Two pagination styles coexist: list endpoints take `limit`/`offset`; the three `*/search`
endpoints take a page-number/size body and return `facets` plus page metadata. Search
field allowlists and closed status enums are enforced in [`wire/search.go`](wire/search.go).

## The OpenAPI → SDK pipeline

The admin surface is the one Janus documents as a machine-readable spec, because a
TypeScript client is generated from it:

```
admin/handlers (ConfigureOpenAPI + All)
  └─ cmd/adminspec  →  web/packages/admin-sdk/data/openapi.json   (make openapi-admin)
       └─ openapi-typescript  →  @janus/admin-sdk  →  web/apps/admin
```

[`cmd/adminspec`](../cmd/adminspec/) builds a bare rocco engine — no DB, DI, or auth,
since endpoint registration only records metadata — and marshals the spec, backfilling one
component schema ([`ValidationFieldError`](../cmd/adminspec/main.go)) that rocco omits. Run
`make openapi-admin` after changing any endpoint's request/response shape.

## Application labels, not IDs

Grant, license, scope, and tier responses show the application by **name**, never its raw
ID, and never via a SQL join. The mapping lives in Redis
([`internal/labels`](../internal/labels/)): `id→name` is always current, `name→id` is
append-only so a renamed app's old name keeps resolving. It's kept warm by
`ApplicationCreated`/`ApplicationUpdated` events plus a boot-time reconcile, and
transformers resolve a batch of IDs in one call.
