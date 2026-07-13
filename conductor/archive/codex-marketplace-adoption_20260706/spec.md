# Spec: Codex Marketplace Adoption

## Overview

Prepare, validate, and submit the OSF Codex plugin to the Codex Marketplace or
record the exact review blocker.

## Functional Requirements

- Validate `.codex-plugin/plugin.json`, skill content, MCP config, marketplace
  metadata, archive paths, and install docs.
- Use public repo materials and existing plugin package, improving any
  marketplace validation score or feedback toward 100%.
- Use Chrome for marketplace submission or login flows; ask the user to log in
  if authentication blocks progress.
- Store receipts, marketplace review queue evidence, score output, or blockers.

## Acceptance Criteria

- Codex plugin is submitted/queued or blocked with exact next action.
- Marketplace/package validators pass locally.
- Go, JSON, anti-stub, review, and registry checks pass.

## Out Of Scope

- Submitting private credentials or private OSF project examples.
