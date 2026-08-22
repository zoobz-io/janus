# Admin contracts

The Go interfaces that define what the admin surface may do. They are the `contracts`
role of the [four-package pattern](../../api/README.md#the-four-package-pattern) — read
that first; this README covers only what is specific to admin.

Same stores as the public API, wider doors. The concrete objects backing these
interfaces are the identical store instances the public API uses — the narrowing is
purely the method set each package declares. Where [`api/contracts`](../../api/contracts/)
exposes the self-service slice (read my own, create my tenant), the admin contracts
expose the operator methods the public interfaces withhold: `CreateApplication` and
`Update` on [`Applications`](applications.go), `Grant`/`UpdateAccess`/`SetTier`/`Revoke`
on [`Grants`](grants.go), `ListAll` regardless of status, `Search` over the full set.
One store, two interfaces; the seam is the interface, not the object behind it.

Eleven interfaces map one-to-one onto stores — `Applications`, `Tenants`,
`Memberships`, `Users`, `Accounts`, `Sessions`, `Licenses`, `Grants`, `Scopes`,
`Tiers`, `Features`. One does not: [`ApplicationLabels`](labels.go) is backed by the
Redis id↔name mapping in [`internal/labels`](../../internal/labels/), not a store. It
exists so responses can show an application's **name** without a SQL join —
`ResolveNames` for outbound id→name in batches, `ResolveID` for inbound name→id when
filtering. See [`transformers/`](../transformers/) for who consumes it.
