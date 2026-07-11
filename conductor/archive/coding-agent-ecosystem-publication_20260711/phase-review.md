# Track 25 Review

Reviewed: 2026-07-11

## Evidence

- Current provider documentation for Cursor, Cline, Roo Code, VS Code, and Zed
  was checked and linked from `docs/coding-agent-ecosystem-evidence.md`.
- Standard MCP templates were added for Cursor, Cline, Roo Code, Windsurf, VS
  Code, and Zed.
- `go run ./tools/checkreleasecontract` validates every template's server
  command, arguments, and environment-only credential references.
- The templates contain no secret values and use the repository's local Go
  development command.

## External boundary

The artifacts are direct MCP configuration integrations. No provider gallery,
one-click collection, or native Zed extension listing is claimed without dated
provider evidence. Zed's native MCP extension route is documented as
deprecated in favor of the official MCP registry.

## Result

The review found and corrected a validator weakness: integration templates are
now required to provide all three credential variables with exact `${env:NAME}`
references. No blocking repository-local findings remain.
