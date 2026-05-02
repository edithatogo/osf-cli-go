# Contributing

## Local Checks

Run the same checks expected by CI before committing:

```powershell
.\scripts\check.ps1
```

On systems with `make`:

```sh
make check
```

For docs and release-readiness work, also run the CLI help path and the anti-stub scan from the repository root:

```powershell
go run ./cmd/osf --help
go run ./tools/checkstubs
```

## Conductor Workflow

- Work from the relevant track under `conductor/tracks/`.
- Do not mark a task complete if production behavior is stubbed or only pretends to work.
- Keep planned command examples clearly labeled as planned until the matching implementation lands.
- At the end of each phase, run `$conductor-review`, apply safe fixes, re-run validation, and write phase review evidence using `conductor/templates/phase-review.md`.

## Status Language

- `scaffolded`: files and shape exist, but behavior is not complete.
- `offline-tested`: behavior is covered by unit tests and fixtures.
- `integration-ready`: behavior is ready for live OSF validation.
- `live-validated`: behavior has passed an explicit live OSF check.

## Live Integration Tests

Live OSF checks must stay opt-in.

- Set `OSF_TOKEN` only for the shell session running the check.
- Use any track-specific test project or node variables only when explicitly running live validation.
- Keep unit tests and fixture-backed checks green without live network access.
- Treat a live run as validation, not as a substitute for offline test coverage.

## Release Docs

Before a release, follow [docs/release-checklist.md](docs/release-checklist.md) and confirm the version, binary matrix, checksums, and validation steps are all complete.
