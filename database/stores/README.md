# Stores

Postgres data access for the [models](../models/). Each store wraps a `sum.Database[T]`
and builds queries with [astql](https://github.com/zoobz-io/astql) — the
[soy](https://github.com/zoobz-io/soy) query builder rendered to SQL — over sqlx. No raw
SQL strings, no ORM. There is no object storage, bucket, or KV layer anywhere in this
package; the only backend is Postgres.

## Constructing the aggregate

```go
func New(db *sqlx.DB, renderer astql.Renderer) *Stores
```

Two arguments — the sqlx handle and an astql renderer (boot passes `astqlpg.New()`) — and
no error return. `New` never fails; it just wires up the individual stores. The `Stores`
struct exposes one field per store:

`Tenants`, `Users`, `Memberships`, `Accounts`, `Sessions`, `Applications`, `Licenses`,
`Grants`, `Scopes`, `Tiers`, `Features`, `Config`.

## Transactions

```go
func (s *Stores) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error
```

Runs `fn` in a single transaction: commit on nil, rollback on error. Store methods with
`Tx` suffixes take the callback's `tx` and participate. Multi-write invariants depend on
it — [`provisioning.go`](provisioning.go) uses it for the flows that must land whole:
`CreateTenantWithOwner` (a tenant never exists without an owner) and `RegisterUser` (a
registered user never exists without a linked account, plus an optional owned tenant in
the same transaction).

## Store methods

Each store exposes typed methods, not a generic CRUD façade. [`Users`](users.go), for
example: `GetUser`, `GetUserByEmail`, `CreateUser` (+ `CreateUserTx`), `UpdateDisplayName`,
`Update`, `TouchLastSeen`, `List`, and `Search`. Read the individual file for a store's
full surface. The admin search stores share one pattern in [`search.go`](search.go): a
filtered page, a total count, and the distinct facet values, all from one WHERE clause —
backed by the trigram indexes in [`migrations/003`, `006`, `007`](../migrations/).
