# Codex Marketplace Submission Evidence

## Audit

- Date: 2026-07-13
- Local provider: Codex CLI plugin marketplace commands.
- Provider documentation: <https://help.openai.com/en/articles/20001256-plugins-in-codex>

## Package Validation

- Source package: `plugins/codex-osf`.
- Marketplace manifest: `.agents/plugins/marketplace.json`.
- Release contract: `go run ./tools/checkreleasecontract` passed.
- The package contains `.codex-plugin/plugin.json`, `.mcp.json`, `README.md`,
  `skills/osf/SKILL.md`, and the release `bin/osf-mcp` archive entry.
- The marketplace source path is `./plugins/codex-osf`, resolved from the
  repository root as required by the Codex CLI marketplace importer.

## Isolated Codex CLI Receipt

With a temporary `HOME` and no user configuration:

```text
codex plugin marketplace add . --json
codex plugin list --available --json
codex plugin add osf-cli-go@osf-cli-go --json
codex plugin list --json
```

Observed results:

- Marketplace added as `osf-cli-go`.
- Available plugin: `osf-cli-go@osf-cli-go`, version `0.3.2`.
- Installation completed into the isolated Codex plugin cache.
- Final listing reported the plugin as installed and enabled.

## Current External State

The Codex CLI provides a repository/Git marketplace import and local
installation path, which is validated above. OpenAI's current documentation
describes the separate Codex Plugin Directory and workspace Plugin settings,
but no public submission form, receipt, directory listing, or approval was
available from the local provider surface during this review.

No public marketplace publication or OpenAI approval is claimed.

## Required Follow-up

Use the OpenAI Plugin Directory/workspace publication workflow when an
authorized publisher or workspace administrator can submit the public plugin.
Record the resulting receipt, listing URL, review state, and any score here.
