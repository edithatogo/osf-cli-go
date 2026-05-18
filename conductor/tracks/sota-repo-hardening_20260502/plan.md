# Plan: SOTA Repository Hardening

## Reconciliation Note

Status reconciled on 2026-05-17 against the current codebase. GitHub templates, dependency automation, local pre-commit config, README badges, module-path migration, docs-site scope, disclosure policy, workspace decision, and phase-review evidence are present.

## Closeout Note

Completed on 2026-05-17 with a manual-dispatch GitHub Pages workflow, MkDocs configuration, documentation landing/development pages, private vulnerability disclosure policy, CI coverage summary/artifact reporting, and an explicit decision not to add `go.work` while the repository remains a single module.

## Phase 1: GitHub Templates

- [x] Task: Create .github/ISSUE_TEMPLATE/bug_report.md
- [x] Task: Create .github/ISSUE_TEMPLATE/feature_request.md
- [x] Task: Create .github/ISSUE_TEMPLATE/config.yml
- [x] Task: Update pull request template with checklist

## Phase 2: Developer Experience

- [x] Task: Add .pre-commit-config.yaml with go fmt, go vet, go test hooks
- [x] Task: Add .github/dependabot.yml for Go module and GitHub Actions security updates
- [x] Task: Add README badges for CI, lint, security, go-reference, license, go-version, release
- [x] Task: Update SECURITY.md with proper disclosure contact
- [x] Task: Add CI coverage summary and artifact reporting

## Phase 3: Documentation Site

- [x] Task: Set up mkdocs-based GitHub Pages site
- [x] Task: Create docs/index.md, docs/install.md, docs/usage.md, docs/development.md
- [x] Task: Add GitHub Actions workflow to build and deploy docs

## Phase 4: Module Path and Workspace

- [x] Task: Update go.mod module path from `osf-cli-go` to `github.com/edithatogo/osf-cli-go`
- [x] Task: Update all internal imports
- [x] Task: Update ldflags in build scripts and .goreleaser.yaml
- [x] Task: Verify build and tests pass with new module path
- [x] Task: Set up Go workspace if beneficial for multi-package development

## Phase 5: Review

- [x] Task: Run quality gates and tests
- [x] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
