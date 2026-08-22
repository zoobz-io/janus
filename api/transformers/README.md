# transformers

Pure functions mapping [`database/models`](../../database/models/) to [`wire`](../wire/)
types. No I/O, no database calls, no mutation of the models passed in — a transformer takes
a model (and whatever context the wire shape needs) and returns a DTO. This is the third
step of [the four-package pattern](../README.md#the-four-package-pattern).

The naming is `XToResponse` for one and `XsToResponse` for a slice:
`SessionToResponse` / `SessionsToResponse`, `AccountToResponse` /
`AccountsToResponse`, and so on. A transformer takes exactly the inputs its wire shape
needs — sometimes just the model, sometimes more:

```go
func UserToResponse(u *models.User, memberships []*models.Membership, tenantNames map[string]string) wire.UserResponse
```

because a `UserResponse` carries the user's memberships with their tenant names, and those
are separate reads the handler gathers first. There is no `Apply*` / in-place-mutation
convention here — everything returns a fresh wire value.
