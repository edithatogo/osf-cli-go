# LobeHub MCP Marketplace adoption

## Overview

Prepare and submit the owned OSF MCP repository to LobeHub Marketplace using
the official market CLI and its GitHub ownership workflow.

## Requirements

- Validate Node.js and `@lobehub/market-cli` prerequisites.
- Prepare `lhm.plugin.json` with release-aligned metadata and tool inventory.
- Use browser authentication only when required; never expose credentials.
- Record listing URL, version, and exact status; do not confuse submission with
  publication or validation.

## Acceptance criteria

- Local manifest and package validation pass.
- Authenticated submission or an exact authentication blocker is recorded.
- Public listing is independently verified before claiming publication.
