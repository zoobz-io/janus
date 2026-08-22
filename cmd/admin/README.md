# cmd/admin

The operator API binary. Serves the admin HTTP surface — operator CRUD over every
application, tenant, user, and access grant — on `config.Admin` port
(`APP_ADMIN_PORT`, default `8081`).

This is the entrypoint; the endpoint list, the OpenAPI description, and the
capability boundary live in [`admin/`](../../admin/). The runtime it boots on lives
in [`internal/boot/`](../../internal/boot/).

## What `run()` does

`main` calls `run() error`. `run` then:

1. **Boots the shared runtime** — [`boot.Init(ctx)`](../../internal/boot/), then
   loads `config.Admin` and `config.Cookie`.
2. **Starts the label cache** — `labels.NewApplicationLabels(rt.Redis, rt.Stores.Applications)`
   maps application id ↔ name in shared Redis; `Start(ctx)` keeps it current via
   domain events and boot reconciliation. See [`internal/labels/`](../../internal/labels/).
3. **Registers the admin contracts** — ApplicationLabels, Applications, Tenants,
   Memberships, Users, Accounts, Sessions, Licenses, Grants, Scopes, Tiers, and
   Features: the *same* shared stores as the public API, narrowed to a broader
   capability boundary. Then `sum.Freeze`.
4. **Wires observability** — OTEL providers plus the aperture bridge.
5. **Wires authentication** — the same cookie + bearer authenticator as the public
   API, over the *same* session store and cookie signing key. Admin runs **no OIDC
   login of its own**: the browser session is minted by the public login flow, and
   its cookie authenticates here because the two surfaces share the store. Service
   callers present `Authorization: Bearer <token>`. See [`internal/auth`](../../internal/auth/).
6. **Serves** — `adminhandlers.ConfigureOpenAPI`, `Handle(adminhandlers.All()...)`,
   then `rt.Svc.Run("", adminCfg.Port)`.

## Run it

```bash
make dev-admin   # admin API + postgres + redis + migrate, in compose
make run-admin   # against an already-running database
```
