# Provider-scoped CLI discovery and inspection

## Objective

Expose reviewed Zenodo read capabilities through explicit provider-scoped CLI
UX while preserving the frozen OSF command contract.

## Requirements

- Prefer provider-scoped commands or an unambiguous selector; do not rename the
  product until multi-provider usage and compatibility justify it.
- Preserve stable JSON, clear help, shell behavior, and actionable errors.
- Reject unsupported operations through capability-aware diagnostics.

## Completion evidence

CLI tests and examples cover human and JSON output, errors, compatibility, and
cross-platform invocation without regressing existing OSF commands.
