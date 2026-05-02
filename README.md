# osf-cli-go

A Go command-line client for the Open Science Framework (OSF).

This project is being set up with Conductor so product intent, technical choices, and delivery workflow stay explicit as the CLI grows.

## Current Status

The CLI is offline-tested for help/version, token inspection, project/component/file listing, and download-safety package behavior. Live OSF validation remains opt-in and write/export commands are still planned in the Conductor tracks.

## Install

Requirements:

- Go 1.26 or newer
- An OSF personal access token when you want to run authenticated commands or live OSF checks

From a local checkout:

```powershell
go build -o bin\osf.exe ./cmd/osf
```

For a local install into your Go bin directory:

```powershell
go install ./cmd/osf
```

## Authentication

Set `OSF_TOKEN` in the shell session that will run OSF commands. Do not commit the token or write it into project files.

PowerShell:

```powershell
$env:OSF_TOKEN = '<your-token>'
```

bash:

```sh
export OSF_TOKEN='<your-token>'
```

## Build And Run

Current commands:

```powershell
go run ./cmd/osf --help
go run ./cmd/osf --version
go run ./cmd/osf auth whoami
go run ./cmd/osf projects list
go run ./cmd/osf projects get <guid-or-url>
go run ./cmd/osf components list <project-guid-or-url>
go run ./cmd/osf files list <project-or-component-guid>
```

After building or installing:

```powershell
osf --help
osf --version
osf auth whoami
```

Planned commands, shown here only as planned examples:

```powershell
osf files download <path>
osf export
```

## Output And Status Language

Use the status words from the Conductor workflow when describing progress:

- `scaffolded`: files and shape exist, but behavior is not complete.
- `offline-tested`: fixture-backed behavior passes without live OSF access.
- `integration-ready`: behavior is ready for an explicit live OSF check.
- `live-validated`: behavior has passed an opt-in live OSF check.

## Safety Defaults

- Commands that write to OSF should be explicit and conservative.
- Public read behavior may work without auth when OSF allows it, but private or account-specific operations should assume `OSF_TOKEN` is required.
- Do not print tokens, persist them in repo-local config, or echo them in logs and failures.
- Live integration tests must stay opt-in and should only use `OSF_TOKEN` and any test fixture variables when you are deliberately running those checks.

## Release Readiness

See [docs/release-checklist.md](docs/release-checklist.md) for the release checklist, versioning policy, binary matrix, checksums, and validation steps.
