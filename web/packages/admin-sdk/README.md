# @janus/admin-sdk

The typed client for the [admin API](../../../admin/). Every path, every request body,
every response is generated from the admin's OpenAPI spec — so the client cannot drift from
the server, and a renamed field is a compile error, not a 2 a.m. surprise.

```ts
import { createAdminClient } from "@janus/admin-sdk";

const admin = createAdminClient();
const { users, total } = await admin.users.list({ query: { limit: "20", offset: "0" } });
const members = await admin.tenants.members.list(tenantId);
```

Methods take positional path params, then a trailing options object; they return the
response body directly and throw from the openapi-press error hierarchy on failure. With no
config, `createAdminClient()` targets the same origin — which is exactly what the browser
wants when a proxy fronts the admin API.

## Generation

The client is generated, never written by hand. Two commands, two halves:

```bash
make openapi-admin   # Go side: dump the live spec → ./data/openapi.json
pnpm generate        # TS side: openapi-typescript ./data/openapi.json -o ./src/schema.ts
```

- **`./data/openapi.json`** is produced on the Go side by
  [`cmd/adminspec`](../../../cmd/adminspec/) — see
  [the admin README](../../../admin/README.md#the-openapi--sdk-pipeline) for how the spec is
  dumped from the handlers.
- **[`src/schema.ts`](src/schema.ts)** is `openapi-typescript` output: **auto-generated —
  do not edit.** Regenerate it after any change to `openapi.json`.

## The client tree

[`src/client.ts`](src/client.ts) uses `definePress<paths>()` to bind each operation to its
generated type, then assembles a resource-namespaced tree — the *only* file here that is
hand-authored:

```
applications  list · create · get · update
  ├─ grants     list · create · update · revoke
  ├─ licenses   list · authorize · revoke
  ├─ scopes     list · create · delete
  └─ tiers      list · create · delete
       └─ features   list · add · remove
providers     list
tenants       list · create · get · update
  └─ members    list · add · updateRole · remove
users         list · create · get · update
  ├─ accounts   list · unlink
  └─ sessions   list · revoke · revokeAll
```

## Exports

From [`src/index.ts`](src/index.ts):

- **`createAdminClient`** and the **`AdminClient`** type (a live, fully-typed client).
- The openapi-press cross-cutting wrappers, applied per endpoint via `.with` —
  `withCache`, `withCircuitBreaker`, `withDedupe`, `withFallback`, `withReauth`,
  `withRetry`, `withTimeout` — so a consumer instruments and hardens calls without a second
  dependency.
- The openapi-press **error hierarchy** (`export * from "openapi-press/error"`) and its
  config/hook types.
- The generated `components` and `paths` types, for naming a request/response shape
  directly (e.g. `components["schemas"]["TenantResponse"]`).
