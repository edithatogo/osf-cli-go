# SOTA Repository Hardening

## Objective

Apply state-of-the-art practices to the GitHub repository setup, CI/CD, automation, documentation, and developer experience.

## Acceptance Criteria

- GitHub issue templates (bug report, feature request, config.yml)
- GitHub pull request template with checklist
- Pre-commit hooks configuration (.pre-commit-config.yaml)
- Dependabot config for security updates (complementing Renovate)
- GitHub Pages documentation site with mkdocs or similar
- README badges: CI, lint, security, go-reference, license, go-version, release
- SECURITY.md with proper disclosure policy
- Go module path updated from local `osf-cli-go` to `github.com/edithatogo/osf-cli-go`
- Go workspace or multi-module setup if beneficial
- Codecov or equivalent coverage reporting

## Non-Goals

- Publishing the documentation site to a custom domain
- Signing or notarizing releases
- Full migration away from Renovate
