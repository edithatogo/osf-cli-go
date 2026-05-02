# osf-cli-go

A Go command-line client for the Open Science Framework (OSF).

[![CI](https://github.com/edithatogo/osf-cli-go/actions/workflows/ci.yml/badge.svg)](https://github.com/edithatogo/osf-cli-go/actions/workflows/ci.yml)
[![Lint](https://github.com/edithatogo/osf-cli-go/actions/workflows/lint.yml/badge.svg)](https://github.com/edithatogo/osf-cli-go/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/edithatogo/osf-cli-go.svg)](https://pkg.go.dev/github.com/edithatogo/osf-cli-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/edithatogo/osf-cli-go)](go.mod)

## Features

- `osf auth whoami` — Identify the authenticated OSF account
- `osf projects list|get` — List and inspect projects
- `osf components list` — List project components
- `osf files list|download` — Browse and download OSF Storage files
- `osf completion bash|zsh|fish|powershell` — Shell completion scripts
- JSON and human-readable output modes
- Safe, atomic file downloads with conflict policy (fail/skip/overwrite)

## Install

Requirements:
- Go 1.26 or newer

```powershell
go install github.com/edithatogo/osf-cli-go/cmd/osf@latest
```

Or from a local checkout:
```powershell
go build -o bin\osf.exe ./cmd/osf
.\scripts\build.ps1
```

## Authentication

Set `OSF_TOKEN` in your shell session. Do not commit the token or write it into project files.

```powershell
$env:OSF_TOKEN = '<your-token>'
```

## Quick Start

```powershell
osf --help
osf auth whoami
osf projects list
osf projects get https://osf.io/abc12/
osf components list abc12
osf files list abc12
osf files download --file <file-id> ./output/
osf files download --tree abc12 ./output/
```

## Output Modes

All commands support `--output table|json` and `--json` shorthand:

```powershell
osf projects list --json
osf auth whoami --output json
```

## Project Status

All 15 Conductor tracks are complete. The CLI is **offline-tested** for all read-only operations including file downloads with conflict handling, path traversal protection, symlink escape prevention, and atomic writes.

## Documentation

- [Release checklist](docs/release-checklist.md)
- [MCP roadmap](docs/mcp-roadmap.md)
- [Contributing](CONTRIBUTING.md)

## License

Apache 2.0 — see [LICENSE](LICENSE).

## Citation

If you use this software in your research, please cite it using the metadata in [CITATION.cff](CITATION.cff).
