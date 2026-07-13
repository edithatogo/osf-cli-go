# OSF CLI Go for Codex

This plugin exposes the read-only OSF MCP server to Codex-compatible hosts.

The release archive bundles the server at `bin/osf-mcp` and loads the MCP
configuration from `.mcp.json`. To install the repository marketplace from the
repository root, run:

```text
codex plugin marketplace add .
codex plugin add osf-cli-go@osf-cli-go
```

For source development, run:

```text
go run ./cmd/osf-mcp
```

Set `OSF_TOKEN` before using authenticated tools. `OSF_USERNAME` and
`OSF_PASSWORD` are supported as a fallback. The Codex CLI does not currently
expose a standalone `plugin validate` subcommand. Installation through an
isolated Codex home is the supported local validation path; the repository
release contract also checks the manifest and source path:

```text
go run ./tools/checkreleasecontract
```
