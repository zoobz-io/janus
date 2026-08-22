# @janus/admin

The operator console — a Nuxt 4 (`4.4.8`) + Vue 3.5 app that fronts the
[admin API](../../../admin/). Eight domain pages, each a table over one admin resource,
gated by a Janus session. No business logic lives here; the app is a typed, authorized
window onto the admin surface.

## What it's built on

It **extends the [`@zoobzio/foundation`](https://www.npmjs.com/package/@zoobzio/foundation)
Nuxt layer** (`extends: ["@zoobzio/foundation"]` in [`nuxt.config.ts`](nuxt.config.ts)) —
that layer supplies the `Workspace` component, the `useTable` factory, and the panel/nav
primitives every page reuses. On top of it, two modules do the heavy wiring:

- **[`@openapi-press/nuxt`](https://www.npmjs.com/package/@openapi-press/nuxt)** — typed
  transport. `press.clients.admin` points at [`shared/presses/admin`](shared/presses/admin.ts),
  which default-exports `createAdminClient` from [`@janus/admin-sdk`](../../packages/admin-sdk/).
  Configured with `host: http://127.0.0.1:8081` and `prefix: /api/admin`, so the browser
  hits the same-origin proxy at `/api/admin` and Nuxt forwards to the admin API.
- **[`@letters-patent/nuxt`](https://www.npmjs.com/package/@letters-patent/nuxt)** — auth.
  `lettersPatent: { contract, login: "/" }` installs the `auth` middleware and the
  `/api/auth/*` handlers that establish and gate the session.

## The eight pages

Every page under [`app/pages/`](app/pages/) is the same shape: a `<Workspace>` bound to a
per-domain composable that builds a `useTable` over `useAdminApi().<domain>.list()`. See
[`composables/users.ts`](app/composables/users.ts) for the pattern in full — a `fetch`
callback that maps table `page`/`pageSize` onto the SDK's `limit`/`offset` query.

Grouped by the same two tags the admin API uses ([`constants/nav.ts`](app/constants/nav.ts)):

- **Catalog** — `applications`, `scopes`, `tiers`, `licenses`, `grants`
- **Directory** — `tenants`, `users`, `providers`

Each page declares its auth in `definePageMeta`:

```ts
definePageMeta({
  layout: "dashboard",
  middleware: "auth",
  auth: { scopes: ["directory:read"] },
});
```

All eight pages currently gate on `directory:read` — the read scope every operator holds.
The heavier contract scopes (`users:manage`, `tenants:manage`, `applications:manage`) are
declared and resolved, ready to gate the mutating flows as they land.

## Auth wiring

The contract is one file. [`shared/contract.ts`](shared/contract.ts) declares the
letters-patent `Contract` — scopes `directory:read`, `users:manage`, `tenants:manage`,
`applications:manage`; roles `operator`, `auditor` — mirroring the seeded `janus-admin`
application. Both the browser transport and the server handlers derive their schema from
this single declaration.

[`server/api/auth/[action].ts`](server/api/auth/[action].ts) builds a
[`@janus/authz`](../../packages/authz/) provider **per request**: it reads the incoming
`cookie` header and forwards it to the Janus public API (`runtimeConfig.janus.authHost`,
`http://127.0.0.1:8080`) resolving authorization for app slug **`janus-admin`**. The
bridge collapses Janus's session into the contract's user — first entitled tenant wins
(an operator acts through one tenant), lifting that tenant's `scopes`, `roles`, and name.

This is the **runtime auth path**, and it is deliberately separate from the data path: auth
resolves against the *public* API on `:8080`, while every table fetch goes through the SDK
to the *admin* API on `:8081`.

## Run it

```bash
pnpm --filter @janus/admin dev     # nuxi dev
pnpm --filter @janus/admin build   # nuxi build
```

Both surfaces must be reachable — the admin API on `:8081` for data, the public API on
`:8080` for auth. From the repo root, `make dev-admin` brings up the admin API and its
dependencies.
