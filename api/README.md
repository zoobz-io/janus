# Public API

The user-facing HTTP surface, served by [`cmd/api`](../cmd/api/). Everything here is
scoped to the caller's own identity: a user reads and edits their profile, manages their
sessions and linked accounts, creates tenants, and reads their own entitlements. Nothing
in this surface lets a user act on another user or configure the mesh — that lives behind
the [admin API](../admin/) and the [mesh API](../internal/mesh/).

## The four-package pattern

This is the canonical description. The [admin surface](../admin/) reuses the same shape;
its README links here rather than repeating it.

A request flows through four packages, each with one job:

| Package | Job |
|---------|-----|
| [`contracts/`](contracts/) | Go interfaces defining what this surface may do. Implemented by [`database/stores`](../database/stores/); a surface sees only the methods its contract declares. |
| [`handlers/`](handlers/) | [rocco](https://github.com/zoobz-io/rocco) endpoints. Each resolves its contract from context (`sum.MustUse`), calls it, and maps the result through a transformer. Handlers orchestrate; they hold no business logic. |
| [`transformers/`](transformers/) | Pure functions mapping [`database/models`](../database/models/) ↔ `wire` types. No I/O. |
| [`wire/`](wire/) | Request/response DTOs. Carry validation ([check](https://github.com/zoobz-io/check)), boundary masking (`OnSend` via [sum](https://github.com/zoobz-io/sum)), and `Clone`. This is the only shape the outside world sees. |

The contract interface is the seam: because handlers depend on an interface, not a
concrete store, the same store instance backs both the narrow public contract and the
broad admin contract — each surface simply declares a different method set.

## Endpoints

All endpoints require an authenticated session. Registered in
[`handlers/handlers.go`](handlers/handlers.go).

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/me` | Get my profile |
| `PUT` | `/me` | Update my profile |
| `POST` | `/me/tenants` | Create a tenant |
| `GET` | `/me/sessions` | List my sessions |
| `DELETE` | `/me/sessions/{id}` | Revoke one session |
| `DELETE` | `/me/sessions` | Revoke all my sessions |
| `GET` | `/me/accounts` | List my linked identity accounts |
| `DELETE` | `/me/accounts/{id}` | Unlink an account |
| `GET` | `/applications` | List available applications |
| `GET` | `/me/tenants/{tenant_id}/applications` | List my application grants within a tenant |
| `GET` | `/me/authorization/{app_slug}` | Get my authorization for one application |

The OIDC login flow adds three more when an authenticator is configured, provided by
[rocco](https://github.com/zoobz-io/rocco)'s session framework: `/auth/login`,
`/auth/callback`, `/auth/logout`.

## Authentication

The authenticator ([`internal/auth`](../internal/auth/)) accepts two credentials, tried
in order:

1. **Session cookie** — minted by the OAuth login flow, signed with the cookie key.
2. **Bearer token** — a mesh-issued session in `Authorization: Bearer <token>`, validated
   against the sessions store.

Either resolves to a `JanusIdentity` carrying the user's ID and email. It deliberately
exposes no tenant, scopes, or roles: a janus user spans many tenants, so the *handler*
determines the tenant for a request and `/me/authorization/{app_slug}` resolves scopes
on demand — the identity itself stays thin.
