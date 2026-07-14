# Phase Review

## Track

- Track: Zenodo sandbox transfers
- Phase: Sandbox proof
- Date: 2026-07-15

## Implemented Behavior

- Ran a disposable 1 MiB transfer against `sandbox.zenodo.org` with a one-use
  `deposit:write` token that excluded publication and email scopes.
- Verified provider-acknowledged upload bytes and MD5, complete download bytes
  and MD5, deterministic checkpoint continuation through a real range request,
  and draft deletion.
- Adapted the client to live bucket receipt fields, integral decimal inventory
  sizes, deposition inventory download links, and neutral content negotiation.
- Removed the shell-local token file and revoked the sandbox token after the
  successful run.

## Anti-Stub Evidence

- `go run ./tools/checkstubs` result: passed.
- Production markers found: none.
- Ignored paths verified: repository anti-stub scanner defaults unchanged.
- Self-scan exclusion verified: scanner tests passed in the full suite.
- Validation evidence link or location: `docs/zenodo-sandbox-validation-evidence.md`.

## Validation Commands

```powershell
go fmt ./...
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
go run ./tools/checkstubs
go run ./tools/checkfeaturematrix
go run ./tools/checkzenodoapi
```

All commands passed on 2026-07-15. The opt-in live harness also passed with
`ZENODO_SANDBOX_VALIDATION=1` and a scoped sandbox credential.

## Conductor Review

- Review command: `$conductor-review zenodo-sandbox-transfers_20260715`.
- Blocking findings: live upload receipts and inventory sizes differed from the
  legacy fixtures; upload receipt metadata was incorrectly treated as a content
  URL; a restrictive download Accept header was rejected by the live service.
- Fixes applied: dual-shape receipt/list parsing, exact integral decimal size
  decoding, post-upload inventory refresh, authoritative download-link use, and
  neutral content negotiation with regression tests.
- Re-review result: no blocking findings; offline and live quality gates passed.

## Status

- Completion claim: live-validated.
- Completion rule: Anti-Stub Evidence is complete and the current branch passed
  `go run ./tools/checkstubs`.
- Residual risks: publication is intentionally unavailable and belongs to #109;
  no Zenodo write command or MCP tool is advertised.
- Next phase: archive #108 and continue to the publication-state track.
