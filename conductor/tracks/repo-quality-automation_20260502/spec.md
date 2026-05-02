# Repo Quality Automation

## Objective

Add SOTA repository management, CI, linting, formatting, validation, coverage, vulnerability scanning, and dependency automation.

## Acceptance Criteria

- GitHub Actions run format, tests, race tests, vet, lint, coverage, vulnerability checks, and docs checks.
- Renovate manages Go modules and GitHub Actions without automerge initially.
- Local commands mirror CI for Windows and common Unix shells.
- CI and docs make it clear that mutation testing and longer fuzzing are scheduled or opt-in until packages exist.
