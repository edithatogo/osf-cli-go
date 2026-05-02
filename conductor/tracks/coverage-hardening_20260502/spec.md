# Spec: Coverage Hardening

Improve coverage where it reduces real regression risk.

## Outcomes

- Add tests for CLI usage errors, unknown commands, redaction behavior, API error parsing, download failure paths, and output helpers.
- Keep tests offline and fixture-backed.
- Prefer targeted branch coverage over chasing aggregate coverage mechanically.
- Record before/after coverage in phase review evidence.

## Non-Goals

- Artificial tests that only execute trivial getters.
- Live OSF tests.
