# wire

The request/response DTOs — the only shape the outside world sees. JSON-tagged structs
that carry three concerns at the boundary: validation, masking, and cloning. This is the
last step of [the four-package pattern](../README.md#the-four-package-pattern).

- **Validation.** Request types implement `Validate()` over
  [check](https://github.com/zoobz-io/check) — e.g. `UpdateProfileRequest.Validate()`
  requires `display_name` and caps it at 255.
- **Boundary masking.** Response types implement `OnSend(ctx)`, which runs the value
  through its [sum](https://github.com/zoobz-io/sum) boundary before marshaling. Sensitive
  fields declare the mask inline with a `send.mask` tag (`send.mask:"email"`,
  `send.mask:"name"` on `UserResponse`), so an email leaves the process masked without the
  handler thinking about it.
- **Clone.** Types implement `Clone()` for a deep copy, so boundary processing never
  mutates a caller's value.

Every response type that masks needs its boundary registered before the registry freezes.
[`boundary.go`](boundary.go) does this in one place:

```go
func RegisterBoundaries(k sum.Key) {
	sum.NewBoundary[UserResponse](k)
	// ... every masked response type
}
```

`RegisterBoundaries` takes the `sum.Key` and returns nothing; call it in the binary's
startup before `sum.Freeze`.
