# internal/auth

Session-based authentication for the janus HTTP surfaces ([public API](../../api/),
[admin](../../admin/)). One question: given a request, who is the user? This package
answers it, or refuses.

## The authenticator

`NewAuthenticator` ([`authenticator.go`](authenticator.go)) builds a rocco-compatible
authenticator that tries two credentials **in order**:

1. **Cookie** — the OAuth login flow's session cookie, via the injected
   `cookieExtractor`. Tried first; on success it wins.
2. **Bearer token** — a mesh-issued session in `Authorization: Bearer <token>`, the
   fallback.

Pass a `nil` cookie extractor and the authenticator is **bearer-only** — the shape the
mesh-issued path uses when there's no login flow. The bearer path is
`sessions.ValidateByToken` → `users.GetUser` → a `JanusIdentity`; a nil session or a
lookup failure is an auth error, not a silent pass.

## The identity

[`identity.go`](identity.go) — `JanusIdentity` implements
[`rocco.Identity`](https://github.com/zoobz-io/rocco) with exactly two fields that mean
anything: `ID()` (the internal user ID) and `Email()`. Everything else is deliberately
inert:

- `TenantID()` returns `""` **on purpose** — a janus user spans many tenants, so the
  *handler* picks the tenant for a request; the identity refuses to guess.
- `Scopes()`/`Roles()` are nil, `HasScope`/`HasRole` are false. The public API doesn't
  authorize off the identity; `/me/authorization/{app_slug}` resolves scopes on demand
  (see [`internal/authz`](../authz/)).

The thinness is the design, not an omission — the same note lives in
[`api/README.md`](../../api/README.md#authentication).

## OIDC discovery

[`discovery.go`](discovery.go) — `Discover(ctx, issuer)` fetches
`{issuer}/.well-known/openid-configuration` and returns the three endpoints janus needs
for the authorization-code flow: `authorization_endpoint`, `token_endpoint`,
`userinfo_endpoint`. It does **not** assume endpoints share the issuer's host —
providers like Google spread them across hosts, which is the whole reason discovery
exists. A document missing any of the three is an error.

## The session store

[`store.go`](store.go) — `SessionStore` implements
[`rocco/session.Store`](https://github.com/zoobz-io/rocco) across two backends:

- **Sessions** live in the janus `sessions` table (`Create`/`Get`/`Delete` over
  `stores.Sessions`; `Get` also rejects expired sessions and joins the user's email).
- **CSRF state tokens** live in Redis, keyed `janus:oauth:state:<state>` with a **10-minute
  TTL** (`stateTTL`). `VerifyState` uses `GetDel` — the token is **single-use**, consumed
  the moment it's checked.

`Refresh` is a **no-op**: the sessions store owns expiry, so there's nothing for the
cookie layer to extend.
