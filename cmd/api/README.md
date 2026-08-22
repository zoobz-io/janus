# cmd/api

The public API binary. Serves the user-facing HTTP surface — profile, sessions,
linked accounts, tenant creation, own entitlements — on `config.App` port
(`APP_PORT`, default `8080`).

This is the entrypoint; the surface it exposes — endpoints, the four-package
pattern, and the auth model — lives in [`api/`](../../api/). The runtime it boots
on lives in [`internal/boot/`](../../internal/boot/).

## What `run()` does

`main` calls `run() error` and lets a returned error be fatal. `run` then:

1. **Boots the shared runtime** — [`boot.Init(ctx)`](../../internal/boot/) returns
   the `Runtime` with the registry left unfrozen.
2. **Loads its own config** — `config.App` and `config.Auth` on top of the shared
   database/redis/encryption/OTEL load.
3. **Registers the public contracts** — Users, Sessions, Accounts, Memberships,
   Tenants, Provisioning, Applications, Licenses, Grants, and Authorizations
   (backed by `authz.NewEntitlements`). Then `wire.RegisterBoundaries` and
   `sum.Freeze` — no registration after this point.
4. **Wires observability** — OTEL providers plus the [aperture](https://github.com/zoobz-io/aperture)
   bridge, then `observe.StartSchemaSync` to publish the event schema.
5. **Wires authentication** — cookie + bearer via `auth.NewSessionStore` and
   `auth.NewAuthenticator`; see [`internal/auth`](../../internal/auth/).
6. **Wires the OAuth login flow** — `auth.Discover` resolves the issuer's OIDC
   endpoints (never assumed to sit under the issuer prefix — Google spreads them
   across hosts), and `handlers.AllWithAuth` mounts `/auth/login`, `/auth/callback`,
   and `/auth/logout`.
7. **Serves** — `rt.Svc.Run("", appCfg.Port)`.

The OAuth `Resolve` callback (`resolveOAuth`) maps IdP userinfo onto a janus user
in three steps — linked account wins, then verified-email linking, then
first-contact registration — documented at the call site in
[`main.go`](main.go).

## Run it

```bash
make dev-api   # API + postgres + redis + migrate, in compose
make run       # against an already-running database
```
