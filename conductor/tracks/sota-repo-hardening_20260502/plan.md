# Plan: SOTA Repository Hardening

## Phase 1: GitHub Templates

- [ ] Task: Create .github/ISSUE_TEMPLATE/bug_report.md
- [ ] Task: Create .github/ISSUE_TEMPLATE/feature_request.md
- [ ] Task: Create .github/ISSUE_TEMPLATE/config.yml
- [ ] Task: Update pull request template with checklist

## Phase 2: Developer Experience

- [ ] Task: Add .pre-commit-config.yaml with go fmt, go vet, go test hooks
- [ ] Task: Add .github/dependabot.yml for Go module and GitHub Actions security updates
- [ ] Task: Add README badges for CI, lint, security, go-reference, license, go-version, release
- [ ] Task: Update SECURITY.md with proper disclosure contact

## Phase 3: Documentation Site

- [ ] Task: Set up mkdocs-based GitHub Pages site
- [ ] Task: Create docs/index.md, docs/install.md, docs/usage.md, docs/development.md
- [ ] Task: Add GitHub Actions workflow to build and deploy docs

## Phase 4: Module Path and Workspace

- [ ] Task: Update go.mod module path from `osf-cli-go` to `github.com/edithatogo/osf-cli-go`
- [ ] Task: Update all internal imports
- [ ] Task: Update ldflags in build scripts and .goreleaser.yaml
- [ ] Task: Verify build and tests pass with new module path
- [ ] Task: Set up Go workspace if beneficial for multi-package development

## Phase 5: Review

- [ ] Task: Run quality gates and tests
- [ ] Task: Run `$conductor-review`, apply fixes, re-run review, and write phase review evidence
