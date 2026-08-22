# testing/benchmarks

The benchmark home. The target is wired; the suite is not written yet.

```bash
make test-bench   # go test -bench=. -benchmem over ./testing/benchmarks/...
```

**There are no benchmarks here yet.** This directory and its `make` target exist so
the first bench has a place to land and a command that already runs it — but until a
`Benchmark…` function is committed, `make test-bench` has nothing to measure. Don't
read a performance story into an empty room.

When you add one, drop a `…_test.go` file here under the `//go:build testing` tag with
a standard `func BenchmarkX(b *testing.B)`, and it's picked up automatically.
