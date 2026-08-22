# .github/workflows

Two workflows. No release pipeline — Janus ships nothing from CI.

## ci.yml

The gate on every push to `main` and every pull request against it. Seven working
jobs plus a gate:

| Job | What it runs |
|-----|--------------|
| **test** | `make test` on Go **1.25** (the module's minimum per `go.mod`). |
| **lint** | [`golangci-lint-action@v7`](https://github.com/golangci/golangci-lint-action), golangci-lint **v2.7.2**, against `.golangci.yml`. |
| **integration** | `make coverage-integration`, then uploads `coverage-integration.out` to Codecov under the **`integration`** flag. |
| **security** | [`gosec@v2.22.11`](https://github.com/securego/gosec) → SARIF, uploaded via `github/codeql-action/upload-sarif@v3`. |
| **coverage** | `make coverage`, then uploads `coverage.out` to Codecov under the **`unit`** flag. |
| **web** | In [`web/`](../../web/): pnpm `typecheck`, `lint`, `inspect`, `test`, `build`. |
| **ci-complete** | Gate job. `needs: [test, integration, lint, security, coverage, web]` — branch protection watches this one. |

## labels.yml

Keeps the repository's label taxonomy in sync with
[`.github/labels.yml`](../labels.yml) via
[`crazy-max/ghaction-github-labeler@v5`](https://github.com/crazy-max/ghaction-github-labeler).

Runs on pushes to `main` that touch the label files (`.github/labels.yml` or the
workflow itself), and on `workflow_dispatch`. **`skip-delete` is on**, so labels
absent from the file — including GitHub's stock set — are left untouched rather than
pruned.
