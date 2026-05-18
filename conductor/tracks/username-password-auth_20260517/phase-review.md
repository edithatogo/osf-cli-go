# Phase Review

## Track

- Track: username-password-auth_20260517
- Phase: closeout
- Date: 2026-05-17

## Implemented Behavior

- Added credential selection in `internal/auth` with three modes:
  anonymous, bearer token, and username/password.
- Preserved `OSF_TOKEN` precedence over `OSF_USERNAME` and `OSF_PASSWORD`.
- Added HTTP Basic request signing for username/password fallback credentials.
- Added `auth login` as a guided token-bootstrap command. It does not claim
  automated OSF token minting because no supported OSF token creation API was
  identified during this pass.
- Updated `auth whoami` output to include active auth mode when available.
- Extended live validation tooling to accept and redact username/password
  credentials.
- Added `auth-research.md` and `auth-capability-matrix.md`.
- Updated README, usage, command, architecture, and release docs.

## Anti-Stub Evidence

- `go run ./tools/checkstubs` result: pass on 2026-05-17.
- Production markers found: none reported by the scanner.
- No password persistence or website scraping was introduced.

## Validation Commands

```powershell
go test ./...
go vet ./...
go run ./tools/checkstubs
go run ./tools/checkreviews
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
go test -race ./...
mkdocs build --strict
```

## Validation Results

- `go test ./...`: pass.
- `go vet ./...`: pass.
- `go run ./tools/checkstubs`: pass.
- `go run ./tools/checkreviews`: pass.
- Coverage total: 76.2%.
- `go test -race ./...`: pass.
- `mkdocs build --strict`: pass.
- Live username/password `auth whoami`: pass on 2026-05-17, with `OSF_TOKEN`
  cleared for the process.
- Live username/password `projects list`: pass on 2026-05-17, with `OSF_TOKEN`
  cleared for the process.
- Live username/password `projects get xj6qc`: pass on 2026-05-17.
- Live username/password `components list xj6qc`: pass on 2026-05-17.
- Live username/password `files list xj6qc`: pass on 2026-05-17.
- Live username/password `files addons xj6qc`: pass on 2026-05-17.
- Live username/password `export xj6qc --json`: pass on 2026-05-17.
- Live `search` and `preprints list` timed out during this pass; these are
  public-context commands and were not counted as username/password auth
  validation results.
- Added CLI support for `files list <node> [folder-id-or-path]` after live validation
  exposed that the lower API layer supported folder segments but the CLI command
  accepted only one positional argument.
- Live username/password `files list xj6qc <folder-id>`: pass on 2026-05-17.
- Live username/password `files list xj6qc <child-folder-id>`: pass on
  2026-05-17 and returned file metadata only.
- Added a 30-second default OSF HTTP timeout after live validation showed public
  search/preprint endpoints could otherwise wait indefinitely.
- Added `--limit` to `search` and `preprints list`, defaulting to 20 records,
  because unbounded pagination reached deep OSF pages and timed out during live
  validation.
- Live bounded `search "medical industrial action" --limit 5 --json`: pass on
  2026-05-17.
- Live bounded `preprints list --limit 5 --json`: pass on 2026-05-17.
- Updated `tools/livevalidation` to include the validated read-only username/password
  surface: identity, project listing, project metadata, components, files, add-ons,
  export, bounded search, and bounded preprint listing. File download remains
  gated on `OSF_VALIDATE_DOWNLOAD`.
- Fixed `tools/livevalidation` so sanitized placeholders are used only for
  evidence/output, not as command execution arguments.
- Live `tools/livevalidation -live -evidence '' -timeout 45s` with
  username/password credentials, `OSF_TOKEN` cleared, and
  `OSF_VALIDATE_PROJECT=xj6qc`: pass on 2026-05-17.

## Conductor Review

- Review command: `$conductor-review`
- Blocking findings: none remaining after local inspection.
- Fixes applied: avoided unsupported token-minting claims, added capability
  matrix, documented live-validation gate, and kept PATs as the preferred API
  automation path.

## Residual Risks

- Direct username/password support is live-validated for account identity,
  project listing, project metadata, component listing, storage listing,
  add-on listing, and export. Other OSF live behavior may vary by account, SSO,
  2FA, and endpoint.
- Automated token creation is not implemented because no supported OSF API or
  OAuth password flow was identified. `auth login` therefore guides the user to
  create a PAT and can format an explicit token export when supplied.

## Status

- Completion claim: complete, with offline tests and live read-only
  username/password validation passed.
- Next phase: none required for this track. Optional future work is live
  write-operation validation against an explicit scratch project.
