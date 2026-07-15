# Product Guidelines

## User-Facing Language

- Use OSF Dictionary terminology for user-facing concepts.
- Prefer "project" and "component" in CLI commands and help text. Use "node" only when exposing API-specific details or advanced diagnostics.
- Be explicit about whether behavior is public unauthenticated access, authenticated read access, or a future write operation.

## Status Honesty

- Do not call a feature finished unless the command is runnable, tested, and documented at the level claimed.
- Use `offline-tested` for fixture-backed behavior, `sandbox-validated` only
  after dated non-production execution with cleanup or retained-resource
  evidence, and `production-validated` only with a dated public production
  receipt. Do not use an unqualified "live-validated" release claim.
- If work is only structural, call it "scaffolded" and leave remaining behavior as explicit pending tasks.

## Safety Defaults

- Never print tokens or private project details in logs, errors, or docs examples.
- Opt-in event logs must redact credentials and local paths, remain separate from machine-readable command output, and document operator-controlled retention.
- Read-only commands may default to concise output. Destructive or write commands must require explicit user intent when added.
- Downloads must fail rather than overwrite unless the user supplies an explicit conflict policy.
