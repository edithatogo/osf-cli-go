# Claude Official Plugin Directory Submission Evidence

## Audit

- Date: 2026-07-13
- Official submission guidance: <https://claude.com/docs/plugins/submit>
- Claude.ai form: <https://claude.ai/settings/plugins/submit>
- Console form: <https://platform.claude.com/plugins/submit>
- Official marketplace: <https://github.com/anthropics/claude-plugins-official>

## Package Validation

- Source package: `plugins/claude-osf`
- Source validation: `claude plugin validate plugins/claude-osf` passed.
- Archive build: `scripts/build-plugin-archives.ps1` passed for all five plugin packages.
- Claude archive contents include `.claude-plugin/plugin.json`, `.mcp.json`, `README.md`, and `bin/osf-mcp`.
- Bundled archive validation: `claude plugin validate dist/plugins/claude-osf-0.3.2-osx-arm64` passed.
- The archive builder now preserves dotfiles and asserts provider-specific manifest, documentation, MCP configuration, and binary entries.

## Current External State

- The Claude.ai submission route redirected to `https://claude.com/logout` in the available Chrome session.
- No authenticated submission form was available.
- No submission, review queue URL, listing URL, receipt, or Anthropic Verified status is claimed.

## Required Follow-up

Authenticate at either official submission form, submit the public repository
URL or the generated Claude ZIP, then record the receipt and review state here.
