# osf-cli-go

A Go command-line client for the Open Science Framework (OSF).

[![CI](https://github.com/edithatogo/osf-cli-go/actions/workflows/ci.yml/badge.svg)](https://github.com/edithatogo/osf-cli-go/actions/workflows/ci.yml)
[![Lint](https://github.com/edithatogo/osf-cli-go/actions/workflows/lint.yml/badge.svg)](https://github.com/edithatogo/osf-cli-go/actions/workflows/lint.yml)
[![Security](https://github.com/edithatogo/osf-cli-go/actions/workflows/security.yml/badge.svg)](https://github.com/edithatogo/osf-cli-go/actions/workflows/security.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/edithatogo/osf-cli-go.svg)](https://pkg.go.dev/github.com/edithatogo/osf-cli-go)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/edithatogo/osf-cli-go)](go.mod)
[![Release](https://img.shields.io/github/v/release/edithatogo/osf-cli-go?include_prereleases)](https://github.com/edithatogo/osf-cli-go/releases)
[![OSF CLI Go MCP server](https://glama.ai/mcp/servers/edithatogo/osf-cli-go/badges/score.svg)](https://glama.ai/mcp/servers/edithatogo/osf-cli-go)

## Features

- `osf auth whoami` — Identify the authenticated OSF account
- `osf auth login` — Guided personal-access-token bootstrap for username/password users
- `osf projects list|get` — List and inspect projects
- `osf components list` — List project components
- `osf files list|download|upload|mkdir|rm` — Browse, download, upload, create folders, and delete OSF Storage files
- `osf search` and `osf preprints list` — Search OSF and list preprints
- `osf registrations create` — Create draft registrations for an existing node
- `osf export` — Export a node snapshot as JSON or a summary table
- `osf-mcp` — Stdio MCP server exposing read-only OSF tools for agent clients
- `osf completion bash|zsh|fish|powershell` — Shell completion scripts
- JSON and human-readable output modes
- Safe, atomic file downloads with conflict policy (fail/skip/overwrite)

## Install

Requirements:
- Go 1.26 or newer

```powershell
go install github.com/edithatogo/osf-cli-go/cmd/osf@latest
go install github.com/edithatogo/osf-cli-go/cmd/osf-mcp@latest
```

Or from a local checkout:
```powershell
go build -o bin\osf.exe ./cmd/osf
go build -o bin\osf-mcp.exe ./cmd/osf-mcp
.\scripts\build.ps1
```

## Authentication

Set `OSF_TOKEN` in your shell session. Do not commit the token or write it into project files. `OSF_USERNAME` and `OSF_PASSWORD` are supported as an opt-in fallback credential source, but personal access tokens remain preferred for automation and for accounts using SSO or two-factor authentication.

```powershell
$env:OSF_TOKEN = '<your-token>'
```

For guided token setup:

```powershell
osf auth login
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
osf files upload --node abc12 ./report.pdf
osf search "open science"
osf preprints list
osf registrations create abc12 --schema <schema-id> --title "Analysis plan"
osf export abc12 --json
```

## MCP Server

`osf-mcp` runs a stdio MCP server with read-only tools:
`osf_whoami`, `osf_projects_list`, `osf_project_get`,
`osf_components_list`, `osf_files_list`, `osf_contributors_list`,
`osf_search`, and `osf_preprints_list`.

Local development configs are included for GitHub Copilot, VS Code, Claude,
Codex, Gemini CLI, and Qwen Code. Public registry metadata is in `server.json`
and `registry/`.

### Related MCP Servers

Other maintained servers in the same Glama portfolio:

- [FYI MCP](https://glama.ai/mcp/servers/edithatogo/fyi-cli) for freedom-of-information request workflows.
- [Healthpoint MCP](https://glama.ai/mcp/servers/edithatogo/healthpoint-rs) for licensed health-service directory data.
- [SourceRight](https://glama.ai/mcp/servers/edithatogo/sourceright) for reference and citation verification.

Related OSF ecosystem servers:

- [OSF connector by pipeworx-io](https://glama.ai/mcp/connectors/io.github.pipeworx-io/osf) for hosted OSF connectivity.
- [Paperclip](https://github.com/matsjfunke/paperclip) for multi-provider scholarly and OSF Preprints search.

## Output Modes

All commands support `--output table|json` and `--json` shorthand:

```powershell
osf projects list --json
osf auth whoami --output json
```

## Project Status

The CLI is **offline-tested** for read-only operations, file downloads, WaterButler write primitives, search, preprint listing, draft registration creation, project create/update/delete operations, and node export. All Conductor tracks are reconciled against their per-track plans with closeout review evidence; live OSF validation remains opt-in because it requires credentials and network access.

## Documentation

- [Release checklist](docs/release-checklist.md)
- [Documentation site source](mkdocs.yml)
- [Install guide](docs/install.md)
- [Usage guide](docs/usage.md)
- [Command reference](docs/commands.md)
- [Examples](docs/examples.md)
- [Architecture](docs/architecture.md)
- [Developer guide](docs/contributing.md)
- [MCP roadmap](docs/mcp-roadmap.md)
- [Contributing](CONTRIBUTING.md)

## License

Apache 2.0 — see [LICENSE](LICENSE).

## Citation

If you use this software in your research, please cite it using the metadata in [CITATION.cff](CITATION.cff).
