# @janus/web

The TypeScript half of Janus: the operator console and the two packages it stands on.
This is a private pnpm workspace — nothing here publishes. It exists to turn the
[admin API](../admin/) into a UI without anyone hand-writing a client or a type.

Three members, each with its own README:

| Member | Package | What it is |
|--------|---------|------------|
| [`apps/admin`](apps/admin/) | `@janus/admin` | The Nuxt 4 operator console — eight domain pages over the admin API. |
| [`packages/admin-sdk`](packages/admin-sdk/) | `@janus/admin-sdk` | The generated, fully-typed admin client. Regenerated from the admin OpenAPI spec. |
| [`packages/authz`](packages/authz/) | `@janus/authz` | The letters-patent auth provider — the runtime login/authorization path to the Janus *public* API. |

## Two paths, not one

The console talks to Janus over two independent surfaces, and keeping them straight is
the whole mental model:

- **Data** — CRUD over applications, tenants, users, grants — goes through
  `@janus/admin-sdk` to the **admin API** (`127.0.0.1:8081`), proxied at `/api/admin`.
- **Auth** — who is signed in and what they may see — goes through `@janus/authz` to the
  **public API** (`127.0.0.1:8080`). Login is redirect-based; the resolved session gates
  the pages. The admin API itself only checks that *a* session exists (see the
  [known gap](../admin/README.md#authentication-and-authorization)).

```
apps/admin ─┬─ @janus/admin-sdk ──/api/admin──▶ admin API   (:8081)   data
            └─ @janus/authz     ──/api/auth───▶ public API  (:8080)   auth
```

## The OpenAPI → SDK pipeline

The admin client is not written; it is generated. The Go half of this pipeline — how the
spec is dumped from the live handlers — is documented in
[the admin README](../admin/README.md#the-openapi--sdk-pipeline). This workspace owns the
TypeScript half:

```
admin/handlers ─▶ cmd/adminspec ─▶ packages/admin-sdk/data/openapi.json   (make openapi-admin)
                                        │
                            openapi-typescript
                                        ▼
                         admin-sdk/src/schema.ts   (generated — do not edit)
                                        │
                       src/client.ts · createAdminClient
                                        ▼
                                   apps/admin
```

Change an admin endpoint's shape and the whole chain re-runs from `make openapi-admin`.

## Layout

```
web/
├── apps/
│   └── admin/          # @janus/admin — the Nuxt console
├── packages/
│   ├── admin-sdk/      # @janus/admin-sdk — generated typed client
│   └── authz/          # @janus/authz — letters-patent auth provider
├── pnpm-workspace.yaml # globs: packages/* and apps/*
├── tsconfig.base.json  # shared compiler options
└── vitest.config.ts    # root coverage config
```

## Toolchain

- **pnpm 10.27.0** (pinned via `packageManager`), **Node ≥ 22** (`engines`).
- Workspace globs `packages/*` and `apps/*`.

## Scripts

Run from `web/`. Everything with `-r` fans out across all three members.

| Script | Does |
|--------|------|
| `pnpm build` | `pnpm -r run build` — builds the SDK packages, then the app. |
| `pnpm dev` | `pnpm -r --parallel run dev` — all members in watch mode. |
| `pnpm lint` | ESLint over the workspace. |
| `pnpm format` / `pnpm inspect` | Prettier write / check. |
| `pnpm test` | `pnpm -r run test` — each member's Vitest suite. |
| `pnpm coverage` | Root Vitest coverage run. |
| `pnpm typecheck` | `pnpm -r run stub && pnpm -r run typecheck`. The stub pass builds `@janus/authz`'s `unbuild` stub so its `.dist` entry points resolve *before* the type pass runs — typecheck never depends on a full build. |

## From the repo root

The [top-level Makefile](../Makefile) drives this workspace so a Go-only checkout needs no
`cd`:

```
make web-install   # pnpm install
make web-check     # pnpm run typecheck
make web-lint      # pnpm run lint
make web-test      # pnpm run test
make web-build     # pnpm run build
make openapi-admin # regenerate admin-sdk/data/openapi.json from the live handlers
```

`make check` and `make ci` fold `web-check`, `web-lint`, `web-test` (and, in CI,
`web-build`) into the wider gate.
