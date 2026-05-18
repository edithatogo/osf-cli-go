# Plan: Username Password Auth

## Phase 1: External Auth Research

- [x] Task: Capture current OSF API authentication guidance from official OSF
  documentation, including PAT, OAuth, username/password, SSO, and 2FA caveats.
- [x] Task: Inspect existing OSF client behavior for username/password request
  signing and document whether it uses HTTP Basic, cookie/session login, or a
  token exchange.
- [x] Task: Determine whether OSF exposes a supported API or OAuth flow for
  creating a bearer token from username/password credentials.
- [x] Task: If an automated token creation path exists, document required
  scopes, token lifetime, revocation behavior, and SSO/2FA failure modes.
- [x] Task: If no automated token creation path exists, document the guided PAT
  fallback and identify the exact OSF token settings URL and scope guidance.
- [x] Task: Decide the first supported implementation mechanism and document
  unsupported mechanisms explicitly.

## Phase 2: Capability Matrix

- [x] Task: Add `auth-capability-matrix.md` covering every current CLI command.
- [x] Task: Classify each command as supported, token-required, public-only,
  unsupported, or pending-live-validation for username/password auth.
- [x] Task: Add a `supported-via-token-bootstrap` classification for commands
  that should run after username/password has produced or accepted a bearer
  token.
- [x] Task: Identify which commands require live OSF validation before they can
  be advertised as username/password supported.

## Phase 3: Token Bootstrap Design

- [x] Task: Design `auth login` or equivalent bootstrap command semantics for
  username/password input, token acquisition, and explicit persistence/export.
- [x] Task: Ensure the bootstrap command never stores passwords and never prints
  tokens unless the user explicitly selects an export/print mode.
- [x] Task: Define storage behavior for acquired tokens, including permissions,
  overwrite prompts, and logout/revocation caveats.
- [x] Task: Define the guided PAT fallback when automated token minting is not
  supported by OSF.

## Phase 4: Credential Model

- [x] Task: Replace token-only internals with a credential type that represents
  anonymous, bearer token, and username/password auth modes.
- [x] Task: Preserve `OSF_TOKEN` precedence over `OSF_USERNAME` and
  `OSF_PASSWORD`.
- [x] Task: Add missing/partial credential errors for username/password auth
  without changing existing missing-token behavior for token-required commands.
- [x] Task: Extend secret redaction to username/password values and HTTP Basic
  authorization headers.

## Phase 5: API Client Request Signing

- [x] Task: Add request signing for username/password auth using the mechanism
  selected in Phase 1.
- [x] Task: Add token-backed request signing for tokens acquired or accepted by
  the bootstrap workflow.
- [x] Task: Keep public unauthenticated reads working with no credentials.
- [x] Task: Add request-header tests for anonymous, bearer token, and
  username/password clients.
- [x] Task: Ensure WaterButler operations use the same credential signing model
  where username/password support is confirmed.

## Phase 6: CLI Surface

- [x] Task: Wire `OSF_USERNAME` and `OSF_PASSWORD` into default CLI client
  construction.
- [x] Task: Add `auth login` or the selected bootstrap command, including
  non-echo password handling if interactive input is implemented.
- [x] Task: Add explicit user choices for token persistence or shell export.
- [x] Task: Update `auth whoami` and the `whoami` alias to report the active auth
  mode without printing secrets.
- [x] Task: Add tests for precedence when token and username/password variables
  are both set.
- [x] Task: Add tests for partial username/password env state.

## Phase 7: Live Validation

- [x] Task: Extend the opt-in live validation tool to accept username/password
  credentials without persisting them.
- [x] Task: Extend live validation to verify token bootstrap or guided PAT
  fallback behavior.
- [x] Task: Run read-only live validation for username/password auth when
  credentials are available, or document the credential-gated skip.
- [x] Task: Run write-operation live validation only against an explicit scratch
  project and only after confirmation, or document the credential-gated skip.
- [x] Task: Store sanitized evidence and update the capability matrix based on
  validated outcomes.

## Phase 8: Documentation

- [x] Task: Update README, usage docs, command docs, architecture docs, and
  release checklist with the final auth contract.
- [x] Task: Document PATs as preferred for automation and username/password as an
  opt-in bootstrap/fallback path with SSO/2FA limitations.
- [x] Task: Document the exact scopes needed for read-only commands,
  write commands, WaterButler writes, and registration creation.
- [x] Task: Ensure examples do not encourage storing passwords in files or
  command history.

## Phase 9: Review And Closeout

- [x] Task: Run `go test ./...`.
- [x] Task: Run `go vet ./...`.
- [x] Task: Run `go run ./tools/checkstubs`.
- [x] Task: Run `go run ./tools/checkreviews`.
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase
  review evidence.
