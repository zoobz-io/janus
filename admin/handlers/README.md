# Admin handlers

The [rocco](https://github.com/zoobz-io/rocco) GET/POST/PATCH/DELETE endpoints for the
admin surface — the `handlers` role of the
[four-package pattern](../../api/README.md#the-four-package-pattern). Each handler
resolves its contract from context with `sum.MustUse`, calls the store, and maps the
result through a [transformer](../transformers/). No business logic lives here; they
orchestrate. The full endpoint table — 33 endpoints, grouped Catalog and Directory —
is in [`../README.md`](../README.md). `All()` in [`handlers.go`](handlers.go) is the
registration list.

Three files carry the shared machinery. [`helpers.go`](helpers.go) has `pathID` (one
path param) and `offsetPage` (builds a `models.OffsetPage` from `limit`/`offset`,
falling back to store defaults). [`errors.go`](errors.go) holds the sentinel errors —
the `Err*NotFound` set plus `ErrLastOwner`. [`openapi.go`](openapi.go)'s
`ConfigureOpenAPI` sets the spec `Info`, the per-tag descriptions, and the
`Catalog`/`Directory` tag groups; it is consumed by
[`cmd/adminspec`](../../cmd/adminspec/) to dump the spec the web SDK is generated from.

**Only the Applications handlers emit events.** `create-application` and
`update-application` in [`applications.go`](applications.go) emit
`ApplicationCreated`/`ApplicationUpdated`; every other handler is a silent store call.
Those two events keep the label cache in [`internal/labels`](../../internal/labels/)
warm — which is what lets the transformers resolve application names instead of IDs.
