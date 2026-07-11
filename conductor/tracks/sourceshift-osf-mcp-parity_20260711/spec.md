# Benchmark SourceShift OSF MCP and close MCP capability gaps

## Overview

Evaluate [SourceShift/osf-mcp-server](https://github.com/SourceShift/osf-mcp-server) as a current OSF tooling reference and ensure OSF CLI Go matches or exceeds every material, maintainable capability. Tracking issue: [#8](https://github.com/edithatogo/osf-cli-go/issues/8).

## Functional Requirements

- Produce a source-backed capability and maturity comparison covering OSF entities, authentication, transfer behavior, automation surfaces, tests, releases, and maintenance.
- Implement each beneficial missing capability through existing CLI, API-client, MCP, and output patterns.
- Add the repository and findings to the maintained competitive comparison table.
- Record explicit rationale for capabilities not adopted.

## Non-Functional Requirements

- Preserve offline deterministic tests, stable JSON, secret redaction, conservative writes, and Windows/macOS/Linux behavior.
- Respect upstream licensing; study behavior and public interfaces without copying incompatible code.
- Keep live OSF validation opt-in.

## Acceptance Criteria

- The comparison is reproducible and dated.
- Every material gap is implemented, split into a follow-up issue, or rejected with rationale.
- Tests, race tests, vet, lint, vulnerability, registry, anti-stub, and review gates pass.
- User documentation and the competitive matrix reflect the resulting behavior.

## Out of Scope

- Reimplementing upstream internals that provide no user-facing advantage.
- Network-dependent unit tests or committing credentials.

