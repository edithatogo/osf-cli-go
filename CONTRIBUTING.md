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

## Conductor Workflow

- Work from the relevant track under `conductor/tracks/`.
- Do not mark a task complete if production behavior is stubbed or only pretends to work.
- At the end of each phase, run `$conductor-review`, apply safe fixes, re-run validation, and write phase review evidence using `conductor/templates/phase-review.md`.

## Status Language

- `scaffolded`: files and shape exist, but behavior is not complete.
- `offline-tested`: behavior is covered by unit tests and fixtures.
- `integration-ready`: behavior is ready for live OSF validation.
- `live-validated`: behavior has passed an explicit live OSF check.
