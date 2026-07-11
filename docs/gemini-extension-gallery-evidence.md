# Gemini CLI Extension Gallery Evidence

Last reviewed: 2026-07-11

## Authoritative requirements

- [Gemini CLI extension releasing guide](https://geminicli.com/docs/extensions/releasing/)
- [Gemini CLI extension reference](https://geminicli.com/docs/extensions/reference/)
- [Gemini CLI extension best practices](https://geminicli.com/docs/extensions/best-practices/)

The releasing guide states that public repositories are automatically indexed
when they have the `gemini-cli-extension` topic and a root
`gemini-extension.json`. The crawler runs daily and listing is conditional on
validation. No manual submission issue or email is required by that guide.

## Public repository readiness

- Repository: <https://github.com/edithatogo/osf-cli-go>
- Root manifest: `gemini-extension.json`
- Root context: `GEMINI.md`
- Discovery topic: `gemini-cli-extension`
- Version: `0.3.1`, aligned with `server.json` and client manifests.
- Self-contained release package: `plugins/gemini-osf`, produced by
  `scripts/build-plugin-archives.ps1` with `gemini-extension.json` at archive
  root and a bundled `bin/osf-mcp` binary.

## Local validation

Gemini CLI 0.41.2 installed the root repository extension and listed it in an
isolated temporary home after explicitly accepting the local-source trust and
third-party consent prompts:

```text
Extension "osf-cli-go" installed successfully and enabled.
✓ osf-cli-go (0.3.1)
MCP servers:
  osf
Settings:
  OSF personal access token: [not set]
  OSF username: [not set]
  OSF password: [not set]
```

The release contract also validates both root and packaged manifests:

```text
go run ./tools/checkreleasecontract
```

## Exact external status

The repository is prepared for automatic gallery discovery and has the required
topic. Gallery indexing is an external crawler result and is not claimed as
visible until a dated gallery listing is observed. No provider approval is
claimed. Credentials were not configured during validation.
