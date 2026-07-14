# Compatibility contract freeze

GitHub issue: [#99](https://github.com/edithatogo/osf-cli-go/issues/99)

## Objective

Convert the prepared OSF schema evidence and existing CLI/MCP contracts into a
reviewed, enforceable 1.0 compatibility boundary.

## Requirements

- Review and pin the OSF API source, license, retrieval date, and commit.
- Document generated versus vendored schema decisions.
- Add golden fixtures for CLI JSON, MCP names/input/output schemas, errors,
  limits, and authentication behavior.
- Run breaking-change detection in CI.
- Document deprecation, migration, and explicitly deferred endpoint rules.
