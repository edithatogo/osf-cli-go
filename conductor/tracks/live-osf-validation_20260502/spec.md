# Spec: Live OSF Validation

Create an opt-in validation path for real OSF API checks without making routine tests network-dependent.

## Outcomes

- A local script runs live validation only when explicit OSF environment variables are present.
- Validation covers `auth whoami`, `projects list`, `projects get`, `components list`, `files list`, and `files download` once the download command exists.
- Evidence output is redactable and safe to attach to Conductor phase reviews.
- Missing environment variables skip cleanly with a clear message.

## Non-Goals

- CI live testing by default.
- Storing OSF tokens or private project identifiers in the repository.
