# Live OSF release validation

GitHub issue: [#97](https://github.com/edithatogo/osf-cli-go/issues/97)

## Objective

Validate the release candidate against a disposable OSF project using scoped
credentials while preserving the repository's opt-in, secret-free boundary.

## Requirements

- Exercise authentication, identity, projects, components, files, search,
  preprints, downloads, uploads, conflicts, cancellation, and MCP tools.
- Use disposable resources for write tests and clean them up safely.
- Record sanitized versions, endpoint results, failures, and cleanup evidence.
- Make the validation repeatable from documented environment variables.
- Record a live-validated result or a precise dated blocker in the launch review.

## Safety

Credentials must never be committed, printed, uploaded, or included in reports.
