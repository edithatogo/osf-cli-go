# Security Policy

## Supported Versions

The project is pre-release. Security fixes are handled on the active default branch.

## Reporting

Do not open public issues containing OSF tokens, private project identifiers, or private research data. Report privately through the repository owner until a formal disclosure channel is configured.

## Local Secret Rules

- Do not commit `OSF_TOKEN` or other credentials.
- Do not write tokens to project-local config.
- Redact tokens from logs, errors, and test output.
- Keep live OSF tests opt-in through explicit environment variables.
