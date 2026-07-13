# Development

This page summarizes the local development contract for contributors. The repository-level [contributing guide](https://github.com/edithatogo/osf-cli-go/blob/main/CONTRIBUTING.md) covers contribution etiquette; this page focuses on validation, documentation, and release-readiness checks.

## Local Setup

```powershell
go mod tidy
```

On Windows, prefer the repository check script because it sets local Go caches under the checkout:

```powershell
.\scripts\check.ps1
```

If the local machine does not have `gcc`, use the explicit local-development escape hatch:

```powershell
.\scripts\check.ps1 -AllowRaceSkip
```

GitHub Actions remains the strict race-test gate.

## Validation Gates

Run these before closing a Conductor task or opening a pull request:

```powershell
go fmt ./...
go test ./...
go vet ./...
go run ./tools/checkstubs
go run ./tools/checkreviews
go run ./tools/checkfeaturematrix
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
```

Run `go test -race ./...` when a C compiler is available.

## Documentation Site

The documentation site uses MkDocs. The docs workflow validates the site on pushes and pull requests. GitHub Pages deployment is intentionally gated behind manual workflow dispatch so repository changes do not publish automatically.

Local preview requires MkDocs:

```powershell
mkdocs serve
```

Build the static site with:

```powershell
mkdocs build --strict
```

## Coverage Reporting

The CI workflow writes `coverage.out`, prints the function coverage table, appends the coverage summary to the GitHub Actions job summary, uploads the coverage profile as a workflow artifact, and publishes the profile to Codecov for project and patch status checks.

## Workspace Decision

This repository remains a single Go module. A `go.work` file is not beneficial while all packages live under one module and the future MCP boundary is still internal. Add a workspace only if a second module is introduced for a real package boundary, generated integration harness, or separately versioned tool.

## Security Rules

- Never commit OSF tokens or private project identifiers.
- Keep live OSF tests opt-in.
- Use redacted fixtures or disposable OSF projects for reproductions.
- Report vulnerabilities through the private disclosure path in the [security policy](https://github.com/edithatogo/osf-cli-go/blob/main/SECURITY.md).
