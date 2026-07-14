# Cline MCP Marketplace submission packet

Status: submitted; provider review and approval are pending.

## Submission fields

- GitHub repository: <https://github.com/edithatogo/osf-cli-go>
- Logo: [`assets/osf-mcp-logo.png`](../assets/osf-mcp-logo.png), 400x400 PNG
- Installation guide: [`llms-install.md`](../llms-install.md)
- Cline configuration: [`integrations/cline/cline_mcp_settings.json`](../integrations/cline/cline_mcp_settings.json)

## Reason for addition

OSF CLI Go gives Cline a focused, read-only interface to Open Science
Framework research projects, components, files, contributors, registrations,
preprints, and search. It is a Go-native MCP server with structured results,
redacted stderr logging, deterministic authentication behavior, release
artifacts, and offline compatibility tests. It expands Cline's research-data
coverage without granting write access to OSF resources.

## Validation evidence

The packet was validated on 2026-07-14 with:

```text
go run ./tools/checkreleasecontract
go test ./...
```

The official submission route is the new-issue flow in
<https://github.com/cline/mcp-marketplace>. No approval is claimed until the
upstream repository records the submission.

## External receipt

Submitted on 2026-07-14: <https://github.com/cline/mcp-marketplace/issues/2024>

The issue is the submission receipt. Approval remains pending until the
official marketplace repository reflects acceptance.
