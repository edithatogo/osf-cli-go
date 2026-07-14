# Compatibility Policy

## Scope

OSF CLI Go follows semantic versioning for its CLI and MCP interfaces.

- Patch releases fix defects without intentionally changing documented behavior.
- Minor releases add backward-compatible commands, flags, fields, and tools.
- Major releases may remove or change documented behavior after a migration note.

JSON output is an automation contract. Existing fields are not renamed or
retyped in a minor release. New fields may be added; consumers should ignore
unknown fields. MCP tool names and required input properties are stable within a
major version.

## Supported environments

The release pipeline builds and tests Linux, macOS, and Windows artifacts. The
minimum Go version is the version declared by `go.mod`; users consume release
artifacts rather than needing Go installed. Live OSF validation remains opt-in.

## Deprecation

Deprecated commands, flags, fields, or tools remain documented for at least one
minor release when practical. Release notes must state the replacement and the
first release in which removal may occur.

## Authentication boundary

`OSF_TOKEN` is the preferred automation credential. Username/password fallback
is intentionally secondary and may not work for SSO or two-factor accounts.
Credentials are never persisted by the project or emitted in logs.

## Enforced 1.0 baseline

The CLI root JSON contract is frozen in
`internal/cli/testdata/compatibility/cli-root.json`. The MCP tool names and
input-property contract is frozen in
`internal/mcpserver/testdata/compatibility/mcp-tools.json`. The corresponding
tests compare generated contracts with these fixtures, and CI runs them in
the explicit `Compatibility contract` step.

The compatibility baseline also covers the documented exit codes, bounded
search and preprint limits, authentication precedence and redaction tests,
and MCP validation-error behavior. A change to a frozen name, required input,
field type, exit code, authentication rule, or limit must update the fixture,
this policy, and `docs/migration-v1.md` in the same change. New fields and
optional inputs remain additive within the 1.x window.

The OSF API source remains a pinned remote manifest in
`docs/osf-api-schema-source.json`. The runtime client is typed and maintained
locally; no generated or vendored replacement schema is accepted until the
source ownership, license, update cadence, and compatibility implications are
reviewed. Endpoint families not represented by the current CLI or MCP
contract remain explicitly deferred in that manifest and the feature matrix.
