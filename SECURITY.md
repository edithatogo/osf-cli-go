# Security Policy

## Supported Versions

The project is pre-release. Security fixes are handled on the active default branch.

## Reporting Vulnerabilities

Do not open public issues containing OSF tokens, private project identifiers, private research data, embargoed project details, or exploitable vulnerability details.

Preferred disclosure channel:

- Use GitHub private vulnerability reporting for this repository: <https://github.com/edithatogo/osf-cli-go/security/advisories/new>

If that channel is unavailable, contact the repository maintainers privately through the maintainer contact details listed on the GitHub repository or owner profile. Public issues and pull requests are appropriate only after the maintainers have confirmed that disclosure is safe.

Please include:

- A concise description of the vulnerability and affected command or package.
- Reproduction steps using redacted data or a disposable OSF project.
- Impact, including whether OSF tokens, private node identifiers, downloads, uploads, or local files are exposed.
- The version, commit, operating system, and Go version used.

## Response Expectations

- The maintainers will acknowledge valid private reports as soon as practical.
- Security fixes are prepared on the active default branch and released with clear upgrade guidance.
- The project will avoid requesting live OSF credentials or private research data unless a minimal reproduction cannot be built otherwise.
- Public disclosure should wait until a fix or mitigation is available, unless immediate public warning is necessary to prevent harm.

## Local Secret Rules

- Do not commit `OSF_TOKEN` or other credentials.
- Do not write tokens to project-local config.
- Redact tokens from logs, errors, and test output.
- Keep live OSF tests opt-in through explicit environment variables.
- Set `OSF_TOKEN` only in the shell session that needs it, and clear it afterward if you used a persistent profile or export command.
