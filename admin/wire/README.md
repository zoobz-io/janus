# Admin wire

The request/response DTOs — the `wire` role of the
[four-package pattern](../../api/README.md#the-four-package-pattern), and the only shape
the outside world sees. Each type carries JSON tags, OpenAPI `description`/`example`
struct tags (so the generated spec documents itself), a `Validate` method
([check](https://github.com/zoobz-io/check)), and `Clone` for deep copies.

[`search.go`](search.go) is the shared search vocabulary, reused by all three
`POST /*/search` endpoints (applications, tenants, users): `DateRange`, `SortSpec`,
`PageRequest`/`PageResponse`, the per-entity field allowlists (which facet, date, and
sort fields are accepted), and the closed status enums that `Validate` enforces. It
lives here, next to the request types that embed it, rather than in a cross-cutting
home — sentinel only records type relationships within the same package, so a shared
location elsewhere would silently drop these component schemas from the generated
OpenAPI spec.
