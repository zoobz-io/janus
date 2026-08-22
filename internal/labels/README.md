# internal/labels

A bidirectional application `id↔name` map in Redis, so API responses carry the human
label instead of a raw application foreign key — and never pay for a SQL join to do it.
Kept current by domain events, reconciled from the source table on boot.

## The two directions

[`applications.go`](applications.go). `Put` writes both, and they age differently:

- `app:id:<id>` → **name**, always current. Overwritten on every rename.
- `app:name:<name>` → **id**, **append-only**. A renamed app's old name keeps
  resolving; a name later reused by a different app repoints to the new id. We only ever
  `Set`, never `Delete`.

Mappings **never expire** (TTL `0`).

## Staying current

`Start(ctx)` subscribes to `ApplicationCreated` and `ApplicationUpdated`
([`events/`](../../events/)) **then** runs `Reconcile` — subscribe-before-reconcile is
deliberate, closing the window where a mutation could slip through the startup gap. It
returns a `stop` func that unsubscribes; on reconcile failure it tears the subscription
back down and returns the error.

`Reconcile` idempotently upserts every application from `apps.ListAll` — cold start,
missed events, out-of-band changes, all covered. An event-handler sync failure isn't
swallowed: it emits a `LabelSyncFailed` warning through capitan.

## Resolving

`ResolveNames(ctx, ids)` is the outbound path: it batches `id→name` in one `MGET`
(dedup'd), and **read-repairs** misses by loading from the applications store and healing
the cache. Ids that resolve to no application are simply omitted — label resolution is
best-effort, never a hard error on a response path.

**Confession:** `ResolveID` (`name→id`, meant for inbound filtering) is defined on the
[admin labels contract](../../admin/contracts/labels.go) and implemented here, but **no
handler calls it yet** — it's exercised only in tests. The prop is on stage; nothing has
asked for it.

## Implementation note

sum ships its KV stores over [grub](https://github.com/zoobz-io/grub)'s `StoreProvider`,
and grub has no redis provider at the pinned version — so
[`redisprovider.go`](redisprovider.go) is a thin adapter from go-redis to
`grub.StoreProvider`. The label store is `sum.NewStore[labelEntry]` rather than a bare
string store because atomization rejects bare strings; `labelEntry` is a one-field struct
wrapping the value.
