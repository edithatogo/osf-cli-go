# Homebrew Tap Adoption Review

## Result

Complete for the project-owned tap publication path.

## Evidence

- Tap: https://github.com/edithatogo/homebrew-osf
- Formula: `Formula/osf-cli-go.rb`
- Release: `v0.3.2`
- Local installation: `brew install edithatogo/osf/osf-cli-go`
- Version smoke test: `osf --version` -> `0.3.2`
- Formula test: `brew test edithatogo/osf/osf-cli-go` passed.
- Source package definition: `packaging/homebrew/osf-cli-go.rb`
- Release assets and checksums: [release v0.3.2](https://github.com/edithatogo/osf-cli-go/releases/tag/v0.3.2)

## Review Validation

- `brew audit --formula edithatogo/osf/osf-cli-go`: passed.
- `brew style --formula edithatogo/osf/osf-cli-go`: passed.
- `brew test edithatogo/osf/osf-cli-go`: passed.
- `go test ./...`, `go test -race ./...`, `go vet ./...`, anti-stub,
  review, registry, release-contract, and `govulncheck` checks: passed.
- Homebrew 6 rejects local-path syntax for `brew audit`; the supported tap
  formula-name invocation above was used instead.

## Remaining external work

Upstream Homebrew/core submission is not required for the project tap route.
The tap remains subject to normal GitHub availability and future formula
maintenance.
