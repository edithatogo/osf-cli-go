# Auth And Public Access

## Objective

Implement token handling and public unauthenticated behavior without leaking secrets.

## Acceptance Criteria

- `OSF_TOKEN` is the primary token environment variable.
- Token values are redacted from errors, logs, and test failures.
- Public read commands can run without auth when OSF permits it.
- `auth whoami` reports the active authenticated OSF account and fails clearly when no token is present.
