# testing

Where the tests that don't sit next to the code live. This directory holds only
this README — the suites are the two subdirectories below.

Most of Janus is covered by Go unit tests that live beside the code they exercise,
as `…_test.go` files under the `//go:build testing` tag. Run them from the repo root:

```bash
make test        # every unit test, race detector on
make test-unit   # the same, -short (skips anything that reaches for Docker)
```

What lives here is everything that can't run beside the code:

| Directory | What it is |
|-----------|------------|
| [`integration/`](integration/) | A separate Go module holding the real integration suite — testcontainers spin up ephemeral Postgres and Redis, the goose migrations run, and the tests drive the actual stores. Needs Docker. |
| [`benchmarks/`](benchmarks/) | Home for `go test -bench` benchmarks, run via `make test-bench`. Wired but still empty — see its README. |

The integration suite is its own module because it pulls in testcontainers and its
Docker toolchain; keeping that dependency tree out of the root `go.mod` is the point.
