# Admin transformers

Pure functions mapping [`database/models`](../../database/models/) ↔
[`wire`](../wire/) — the `transformers` role of the
[four-package pattern](../../api/README.md#the-four-package-pattern). No I/O beyond the
label lookups noted below. Timestamps go out as RFC 3339. Alongside the mapping,
`Resolve*Search` (for applications, tenants, users) applies the search contract
defaults — page 1, size 25, sort `updated_at desc` — in one place and builds the
`stores.*SearchParams`, returning the resolved page number and size so the response
transformer can compute page metadata.

What is distinctive here is label resolution. Four transformers take a
[`contracts.ApplicationLabels`](../contracts/labels.go) and resolve application IDs to
**names in a single batch** so responses carry the app name, never its raw ID:
`GrantsToResponse`, `LicensesToResponse`, `ScopesToResponse`, `TiersToResponse` (each
with a singular sibling). The names come from the Redis mapping in
[`internal/labels`](../../internal/labels/) — no SQL join — kept current by the
`ApplicationCreated`/`ApplicationUpdated` events the Applications handlers emit.
