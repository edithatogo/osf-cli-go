# Username Password Auth

## Objective

Add a safe, explicitly scoped username/password authentication path alongside the
existing `OSF_TOKEN` personal access token flow, with a preferred bootstrap
workflow that converts a successful username/password sign-in into a token-backed
CLI session when OSF provides a supported mechanism.

This track must not assume that username/password authentication is equivalent
to bearer-token authentication for every OSF API or WaterButler operation, and
must not assume that OSF exposes a supported API for creating personal access
tokens from a password. Implementation must first identify which current CLI
functions can be exercised with username/password credentials, which require
bearer tokens, which can be unlocked by a credential-to-token bootstrap workflow,
and which are blocked by OSF account settings such as institutional SSO or
two-factor authentication.

## Context

- Current repo behavior is token-only: `OSF_TOKEN` is loaded and sent as
  `Authorization: Bearer <token>`.
- OSF support documentation describes personal access tokens as the API-oriented
  alternative to password authentication.
- OSF user-facing documentation describes creating personal access tokens through
  Account Settings; this track must verify whether a supported programmatic token
  creation endpoint exists before implementing automated token minting.
- Existing OSF ecosystem tooling, including `osfclient`, supports either
  `OSF_TOKEN` or `OSF_USERNAME` plus `OSF_PASSWORD` for protected project/file
  access.
- Password credentials are higher-risk than scoped tokens and must never be
  stored by default or echoed in command output, logs, test failures, live
  validation evidence, or generated docs examples.

## Bootstrap Workflow Contract

The preferred outcome is:

1. User provides `OSF_USERNAME` and `OSF_PASSWORD`; a future interactive
   password prompt must use non-echo input before it can be documented as
   supported.
2. The CLI verifies the account with OSF using a documented and supported
   mechanism.
3. If OSF exposes a supported API or OAuth flow for token creation, the CLI
   obtains a bearer token with the narrowest sufficient scope.
4. The CLI stores or exports the token only through an explicit user-approved
   target, such as printing a shell command, writing a protected local auth file,
   or updating the current process environment where possible.
5. All subsequent API and WaterButler calls use the token-backed path.

If OSF does not expose a supported automated token-generation mechanism, the
accepted fallback is a guided workflow:

- verify username/password where possible,
- open or print the OSF token settings URL,
- tell the user which token scope is required for the requested command set,
- accept the generated token through `OSF_TOKEN` or an explicit `auth token set`
  flow,
- never scrape the OSF website or automate browser password entry unless the
  plan is updated with a separate security review and user confirmation.

## Credential Contract

- `OSF_TOKEN` remains the preferred and highest-precedence authentication method.
- `OSF_USERNAME` and `OSF_PASSWORD` provide an opt-in fallback only when
  `OSF_TOKEN` is absent.
- A token created or supplied through the bootstrap workflow must be treated the
  same as `OSF_TOKEN` for command authorization after acquisition.
- Missing or partial username/password credentials must fail with a clear
  redacted error.
- Password input may be supported interactively in a later phase, but the first
  implementation should support environment variables only unless the plan is
  updated with terminal-input tests and non-echo behavior.
- The CLI must expose which auth mode is active without printing secret values.
- Credential redaction must cover:
  - bearer tokens,
  - `OSF_TOKEN=<value>`,
  - `OSF_USERNAME=<value>` when paired with credential errors,
  - `OSF_PASSWORD=<value>`,
  - HTTP Basic `Authorization` headers,
  - username/password values embedded in error strings.

## Capability Contract

Before implementation is marked complete, produce and maintain an auth capability
matrix covering at least the current command surface:

- `auth whoami`
- `projects list`
- `projects get`
- `projects create`
- `projects update`
- `projects delete`
- `components list`
- `files list`
- `files download`
- `files upload`
- `files mkdir`
- `files rm`
- `files addons`
- `search`
- `preprints list`
- `registrations create`
- `export`

Each row must classify username/password support as one of:

- `supported-offline-fixture`
- `supported-live-validated`
- `supported-via-token-bootstrap`
- `token-required`
- `public-only`
- `unsupported-by-osf`
- `unknown-pending-live-validation`

Unknown rows cannot be described as implemented behavior in user docs.

## Acceptance Criteria

- An auth abstraction supports bearer-token and username/password request
  signing without leaking concrete credential values.
- A bootstrap design is implemented when OSF exposes a supported token creation
  path, or explicitly documented as unavailable with a guided PAT fallback when
  no supported path exists.
- `OSF_TOKEN` behavior remains backward compatible and takes precedence over
  username/password credentials.
- Environment-backed username/password credentials are wired into CLI client
  construction.
- Unit tests cover credential precedence, missing/partial credentials, redaction,
  request headers, and authenticated command behavior for both auth modes where
  fixture support is claimed.
- The capability matrix is committed with evidence for every command listed
  above.
- Live validation is opt-in and records sanitized evidence only.
- Documentation states that PATs are preferred for API automation and explains
  the limitations of username/password auth, token bootstrap, and SSO/2FA
  caveats.
- `$conductor-review`, `go test ./...`, `go vet ./...`,
  `go run ./tools/checkstubs`, and `go run ./tools/checkreviews` pass before
  the track can be marked complete.
