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
