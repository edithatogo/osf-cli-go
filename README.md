# osf-cli-go

A Go command-line client for the Open Science Framework (OSF).

This project is being set up with Conductor so product intent, technical choices, and delivery workflow stay explicit as the CLI grows.

## Current Status

The initial scaffold provides a small `osf` command with version and help output. The first implementation track is planned in `conductor/tracks/mvp-osf-readonly-cli_20260502/plan.md`.

## Development

```powershell
go test ./...
go run ./cmd/osf --help
```
