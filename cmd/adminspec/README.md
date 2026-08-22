# cmd/adminspec

Dumps the admin API's OpenAPI spec to disk. No server, no database — this is a CLI
that feeds the [`@janus/admin-sdk`](../../web/packages/admin-sdk/) client generator.
See the admin surface's [OpenAPI → SDK pipeline](../../admin/) for where the spec
goes next.

## What `run()` does

1. **Builds a bare engine** — `rocco.NewEngine()` carries no DB, DI, or auth.
   Endpoint registration records handler metadata only (paths, request/response
   types, error defs), and each handler resolves its dependencies lazily at request
   time — so `GenerateOpenAPI` needs no runtime boot.
2. **Applies the admin handler set** — `adminhandlers.ConfigureOpenAPI(e)` and
   `e.WithHandlers(adminhandlers.All()...)`: the exact metadata and endpoints the
   admin server serves.
3. **Generates and patches** — `e.GenerateOpenAPI(nil)`, then `patchSpec` backfills
   one component schema, `ValidationFieldError`, that rocco v0.1.19 emits a `$ref`
   to but never adds to `components`.
4. **Writes indented JSON** — to the output path, creating the directory as needed.

## Output path

Defaults to `web/packages/admin-sdk/openapi.json`; override with the first argument.
The Makefile passes the SDK's data snapshot:

```bash
make openapi-admin   # go run ./cmd/adminspec web/packages/admin-sdk/data/openapi.json
```
