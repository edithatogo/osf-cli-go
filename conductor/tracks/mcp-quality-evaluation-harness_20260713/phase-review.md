# MCP quality harness phase review

Reviewed 2026-07-14.

## Offline contracts

- Inventory checks cover the MCPB tool set, descriptions, sensitive auth
  configuration, registry tool inventory, supported client configuration maps,
  server redaction/error handling, and bounded limits.
- Existing in-process MCP server tests remain the behavioral source for
  structured outputs, schema properties, tool errors, and redaction.

## Compatibility and live mode

- `go run ./tools/checkmcpquality` produces a versioned JSON report with named
  checks and pass/fail/skip counts.
- Supported client configuration shapes are parsed offline.
- Live OSF validation is opt-in through the existing credential-gated
  `tools/livevalidation` command and never writes credentials to the report.

## Closeout

- Release and registry checks require `docs/mcp-quality-report.json` to match
  the current package version and have a passing status.
- The feature matrix output contract now records the compatibility harness as
  implemented.
- The committed report records six passing offline checks and one explicitly
  skipped live check because live credentials were not supplied.
