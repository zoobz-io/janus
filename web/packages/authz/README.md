# @janus/authz

The letters-patent auth **provider** for Janus. An app owns a
[letters-patent](https://www.npmjs.com/package/letters-patent) contract — its scopes,
roles, and user shape. This package is the vendor behind that contract: it turns the
contract into live login and authorization backed by the Janus **public** auth API.

This is the runtime **auth** path. It is _not_ the admin data SDK
([`@janus/admin-sdk`](../admin-sdk/)) — that one talks CRUD to the admin API; this one talks
identity and entitlements to the public API. An app uses both.

```ts
import { defineSchema } from "letters-patent";
import { createProvider } from "@janus/authz";

const provider = createProvider(
  schema,
  { api: "/api/janus", app: "janus-admin" },
  (session) => ({
    id: session.identity.id,
    email: session.identity.email,
    name: session.identity.display_name,
    scopes: session.authorization.tenants[0]?.scopes ?? [],
    roles: session.authorization.tenants[0]?.roles ?? [],
  }),
);
```

## What the provider does

`createProvider = defineProvider<Options, Session>` ([`src/provider.ts`](src/provider.ts))
yields three letters-patent callbacks:

- **`login`** — redirect-based. In a browser it navigates to `{api}/auth/login`; the
  session lands in a cookie on return. Server-side callers navigate to `loginUrl(api)`
  themselves. No state is touched here — `resolve` picks the session up.
- **`logout`** — GETs `/auth/logout` and clears the current session.
- **`resolve`** — the real work. It fetches `/me` (identity) and
  `/me/authorization/{app}` (app-scoped authorization) **in parallel**, both carrying the
  ambient cookie (`credentials: "include"`). A `401` on either → signed out. And unless
  `requireEntitlement` is `false`, a valid session with **no entitled tenants** also
  resolves to signed out — being authenticated to Janus but not entitled to _this_ app
  means not signed in _for this app_. Otherwise it hands `{ identity, authorization }` to
  the app-supplied bridge.

Server-side callers (e.g. a Nuxt crest handler) pass a `headers` carrying the incoming
request's cookie so the forwarded fetch authenticates as the browser.

## Effective scopes

An `AuthorizedTenant.scopes` is the tenant's **effective** scope set: the grant's explicit
scopes unioned with the scopes bundled into its tier's features. Janus computes this
server-side — this package receives it already resolved — mirroring the resolver in
[`internal/authz`](../../../internal/authz/). Edit a tier and the next `resolve` reflects it.

## Exports

From [`src/index.ts`](src/index.ts):

- **`createProvider`** and **`loginUrl`**.
- The Janus wire types — `Application`, `Authorization`, `AuthorizedTenant`, `Identity`,
  `Membership`, `Options`, `Session` ([`src/types.ts`](src/types.ts)). These are Janus's
  vocabulary, never an app contract's.

## Build

Built with [`unbuild`](https://www.npmjs.com/package/unbuild) to `.dist` (ESM + `.d.ts`;
see [`build.config.ts`](build.config.ts)):

```bash
pnpm build   # unbuild
pnpm stub    # unbuild --stub — writes .dist stubs that point back at src
```

Because `package.json` resolves its exports through `.dist`, the workspace typecheck runs
`stub` before typechecking so those entry points exist without a full build. See the
[workspace README](../../README.md#scripts).
