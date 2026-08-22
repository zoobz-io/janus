# cmd/mesh

The [aegis](https://github.com/zoobz-io/aegis) gRPC mesh node — the surface peer
services call to resolve identities, mint and validate sessions, read the
directory, and manage entitlements. Listens on `config.Mesh` `Addr()`
(`APP_MESH_HOST`:`APP_MESH_PORT`, default `0.0.0.0:9090`).

This is the entrypoint. The four gRPC services it registers — Identity, Session,
Directory, Entitlement — are implemented in [`internal/mesh/`](../../internal/mesh/).

## What `run()` does

`main` calls `run() error`. `run` then:

1. **Boots the shared runtime** — [`boot.Init(ctx)`](../../internal/boot/), loads
   `config.Mesh`, and freezes the registry immediately. The mesh servers hold
   concrete stores directly, so no contracts are registered — but the freeze is
   still required for store queries to run.
2. **Bootstraps sctx auth** — an [sctx](https://github.com/zoobz-io/sctx) admin from
   a file-based keychain (`aegis.NewFileKeychain` over `APP_MESH_CERT_DIR`), then a
   node self-token, then one scoped `sctx.Guard` per capability
   (`identity:resolve`, `session:manage`, `directory:write`, `entitlement:manage`,
   and the rest).
3. **Builds the node** — `aegis.NewNodeBuilder` wires each gRPC method to its guard
   and registers all four service implementations.
4. **Serves and blocks** — starts the server, then waits on `SIGINT`/`SIGTERM`
   before a clean shutdown.

## Not in compose

The mesh node is **not** wired into `docker-compose.yml`. It needs the sctx
keychain under `APP_MESH_CERT_DIR`, which the compose dev stack does not provision.
Run it against an already-running database:

```bash
make run-mesh
```
