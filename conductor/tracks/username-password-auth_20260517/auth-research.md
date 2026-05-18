# Auth Research

Date: 2026-05-17

## Findings

- OSF support guidance describes personal access tokens as the API/script
  automation path and directs users to create tokens from Account Settings.
- I did not identify a documented OSF API v2 endpoint or OAuth password grant
  that creates a personal access token from username/password credentials.
- Existing ecosystem clients such as `osfclient` document `OSF_USERNAME` and
  `OSF_PASSWORD` as an alternative to `OSF_TOKEN`; the compatible request
  signing mechanism implemented here is HTTP Basic auth.
- SSO and 2FA may prevent password-based authentication even when a bearer token
  works.

## Implementation Decision

- Keep `OSF_TOKEN` as the preferred and highest-precedence mechanism.
- Add `OSF_USERNAME` and `OSF_PASSWORD` as an opt-in fallback credential mode
  using HTTP Basic request signing.
- Add `auth login` as a guided token-bootstrap command rather than claiming
  automated token minting.
- If OSF later documents an automated token creation path, `auth login` can be
  extended to exchange username/password for a token and then reuse the existing
  bearer-token path.

## Unsupported Mechanisms

- No website scraping for token creation.
- No browser automation for password entry.
- No password persistence.
- No claim that username/password works for every endpoint until live validation
  evidence confirms it.
